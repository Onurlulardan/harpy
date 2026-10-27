-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION notifications_v1_adoption_project_v1(
    p_operation_id uuid,p_run_id uuid,p_scope_id uuid
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_kind text;
    v_stage notifications_v1_adoption_stage%ROWTYPE;
    v_final_lsn pg_lsn;
    v_generation bigint;
    v_source_count bigint;
    v_target_count bigint;
    v_actual_count bigint;
    v_source_sha text;
    v_target_sha text;
    v_queue_sha text;
    v_provider_sha text;
    v_legacy_sha text;
    v_bridge_sha text;
    v_receipt_sha text;
    v_rows bigint;
    v_now timestamptz;
BEGIN
    v_now := pg_catalog.transaction_timestamp();
    v_final_lsn := notifications_v1_adoption_projection_ready_v1(
        p_operation_id,p_run_id,p_scope_id
    );
    SELECT operation.cutover_generation INTO STRICT v_generation
    FROM platform.cutover_operations AS operation
    WHERE operation.operation_id=p_operation_id AND operation.run_id=p_run_id;
    FOREACH v_kind IN ARRAY ARRAY[
        'message','effect','delivery','attempt','receipt'
    ]::text[]
    LOOP
        FOR v_stage IN
            SELECT stage.* FROM notifications_v1_adoption_stage AS stage
            WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
              AND stage.target_kind=v_kind AND NOT stage.removed
            ORDER BY stage.target_identity
        LOOP
            CASE v_kind
                WHEN 'message' THEN
                    PERFORM notifications_v1_adoption_project_message_v1(
                        v_stage.target_record
                    );
                WHEN 'effect' THEN
                    PERFORM notifications_v1_adoption_project_effect_v1(
                        v_stage.target_record
                    );
                WHEN 'delivery' THEN
                    PERFORM notifications_v1_adoption_project_delivery_v1(
                        v_stage.target_record
                    );
                WHEN 'attempt' THEN
                    PERFORM notifications_v1_adoption_project_attempt_v1(
                        v_stage.target_record
                    );
                WHEN 'receipt' THEN
                    PERFORM notifications_v1_adoption_project_receipt_v1(
                        v_stage.target_record
                    );
            END CASE;
        END LOOP;
    END LOOP;
    IF EXISTS (
        SELECT 1 FROM notifications_v1_adoption_stage AS stage
        WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
          AND NOT stage.removed AND (
            stage.target_kind='message' AND NOT EXISTS (
                SELECT 1 FROM notifications_messages AS target
                WHERE target.id=stage.target_id
                  AND pg_catalog.to_jsonb(target)=stage.target_record
            ) OR stage.target_kind='effect' AND NOT EXISTS (
                SELECT 1 FROM notifications_effect_receipts AS target
                WHERE target.message_id=stage.target_id
                  AND pg_catalog.to_jsonb(target)=stage.target_record
            ) OR stage.target_kind='delivery' AND NOT EXISTS (
                SELECT 1 FROM notifications_deliveries AS target
                WHERE target.id=stage.target_id
                  AND pg_catalog.to_jsonb(target)=stage.target_record
            ) OR stage.target_kind='attempt' AND NOT EXISTS (
                SELECT 1 FROM notifications_delivery_attempts AS target
                WHERE target.id=stage.target_id
                  AND pg_catalog.to_jsonb(target)=stage.target_record
            ) OR stage.target_kind='receipt' AND NOT EXISTS (
                SELECT 1 FROM notifications_delivery_receipts AS target
                WHERE target.id=stage.target_id
                  AND pg_catalog.to_jsonb(target)=stage.target_record
            )
          )
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Notifications V1 projected target differs from signed source';
    END IF;
    SELECT pg_catalog.count(*),pg_catalog.encode(pg_catalog.sha256(
        pg_catalog.convert_to(COALESCE(pg_catalog.string_agg(
            source_identity||':'||source_sha256,',' ORDER BY source_identity),''),'UTF8'
        )),'hex') INTO v_source_count,v_source_sha
    FROM (
        SELECT stage.source_identity,
            pg_catalog.min(pg_catalog.encode(stage.source_sha256,'hex')) AS source_sha256
        FROM notifications_v1_adoption_stage AS stage
        WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
          AND NOT stage.removed GROUP BY stage.source_identity
    ) AS source;
    SELECT pg_catalog.count(*),pg_catalog.encode(pg_catalog.sha256(
        pg_catalog.convert_to(COALESCE(pg_catalog.string_agg(
            stage.target_kind||':'||stage.target_identity||':'||
            pg_catalog.encode(stage.target_sha256,'hex'),',' ORDER BY
            stage.target_kind,stage.target_identity),''),'UTF8')),'hex')
    INTO v_target_count,v_target_sha
    FROM notifications_v1_adoption_stage AS stage
    WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
      AND NOT stage.removed;
    SELECT (SELECT pg_catalog.count(*) FROM notifications_messages)+
        (SELECT pg_catalog.count(*) FROM notifications_effect_receipts)+
        (SELECT pg_catalog.count(*) FROM notifications_deliveries)+
        (SELECT pg_catalog.count(*) FROM notifications_delivery_attempts)+
        (SELECT pg_catalog.count(*) FROM notifications_delivery_receipts)
    INTO v_actual_count;
    IF v_actual_count IS DISTINCT FROM v_target_count THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Notifications V1 projected target count differs';
    END IF;
    v_queue_sha := notifications_v1_adoption_queue_evidence_v1(
        p_operation_id,p_run_id,p_scope_id
    );
    v_provider_sha := notifications_v1_adoption_provider_evidence_v1(p_scope_id);
    v_legacy_sha := notifications_v1_adoption_legacy_evidence_v1(
        p_operation_id,p_run_id,p_scope_id
    );
    v_bridge_sha := notifications_v1_adoption_bridge_disabled_v1(p_scope_id);
    v_receipt_sha := pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
        'harpy/notifications/v1-cutover-receipt/v1:'||p_operation_id::text||':'||
        p_run_id::text||':'||p_scope_id::text||':'||v_generation::text||':'||
        v_source_count::text||':'||v_target_count::text||':'||v_source_sha||':'||
        v_target_sha||':'||v_queue_sha||':'||v_provider_sha||':'||v_legacy_sha||':'||
        v_bridge_sha||':'||v_final_lsn::text,'UTF8'
    )),'hex');
    INSERT INTO notifications_v1_cutover_receipts(
        operation_id,run_id,tenant_id,cutover_generation,source_count,target_count,
        source_root_sha256,target_root_sha256,queue_evidence_sha256,
        provider_evidence_sha256,legacy_evidence_sha256,bridge_state,final_lsn,
        projected_at,receipt_sha256
    ) VALUES (
        p_operation_id,p_run_id,p_scope_id,v_generation,v_source_count,v_target_count,
        v_source_sha,v_target_sha,v_queue_sha,v_provider_sha,v_legacy_sha,
        'disabled_missing_signed_mtls_contract',v_final_lsn,v_now,v_receipt_sha
    );
    UPDATE notifications_writer_authority SET writer='v2',lifecycle_state='writable',
        generation=v_generation,row_version=row_version+1,updated_at=v_now
    WHERE singleton AND generation<v_generation AND (
        writer='unassigned' AND lifecycle_state='bootstrap' OR
        writer='v1' AND lifecycle_state IN ('quiescing','fenced')
    );
    GET DIAGNOSTICS v_rows=ROW_COUNT;
    IF v_rows IS DISTINCT FROM 1::bigint THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Notifications V1 writer authority switch differs';
    END IF;
    RETURN v_target_sha;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION notifications_v1_adoption_project_v1(
    uuid,uuid,uuid
) FROM PUBLIC;
