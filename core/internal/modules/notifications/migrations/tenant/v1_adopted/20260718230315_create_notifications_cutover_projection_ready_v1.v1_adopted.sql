-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION notifications_v1_adoption_projection_ready_v1(
    p_operation_id uuid,p_run_id uuid,p_scope_id uuid
)
RETURNS pg_lsn
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_expected bigint;
    v_complete bigint;
    v_final_count bigint;
    v_final_lsn pg_lsn;
    v_existing bigint;
    v_assert text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Notifications V1 projection principal denied';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM tenancy_schema_identity AS identity
        WHERE identity.singleton AND identity.tenant_id=p_scope_id
          AND identity.schema_name=pg_catalog.current_schema()
          AND identity.origin='v1_adopted'
    ) OR NOT EXISTS (
        SELECT 1 FROM platform.cutover_operations AS operation
        WHERE operation.operation_id=p_operation_id AND operation.run_id=p_run_id
          AND operation.scope_kind='tenant' AND operation.scope_id=p_scope_id
          AND operation.group_id='notifications' AND operation.domain='notifications'
          AND operation.state IN ('planned','copying','shadow','quiescing')
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Notifications V1 projection authority is not ready';
    END IF;
    SELECT pg_catalog.count(*) INTO v_expected
    FROM platform.cutover_operation_checkpoint_keys AS expected
    WHERE expected.operation_id=p_operation_id AND expected.run_id=p_run_id
      AND expected.checkpoint_key='notification';
    SELECT pg_catalog.count(*),pg_catalog.count(DISTINCT checkpoint.final_lsn::text)
    INTO v_complete,v_final_count
    FROM platform.cutover_copy_checkpoints AS checkpoint
    WHERE checkpoint.operation_id=p_operation_id AND checkpoint.run_id=p_run_id
      AND checkpoint.state='complete' AND checkpoint.final_lsn IS NOT NULL;
    IF v_expected IS DISTINCT FROM 1::bigint OR
       (SELECT pg_catalog.count(*)
        FROM platform.cutover_operation_checkpoint_keys AS expected
        WHERE expected.operation_id=p_operation_id AND expected.run_id=p_run_id)
       IS DISTINCT FROM 1::bigint OR v_complete IS DISTINCT FROM v_expected OR
       v_final_count IS DISTINCT FROM 1::bigint OR EXISTS (
        SELECT 1
        FROM platform.cutover_operation_checkpoint_keys AS expected
        LEFT JOIN platform.cutover_copy_checkpoints AS checkpoint
          ON checkpoint.operation_id=expected.operation_id
         AND checkpoint.run_id=expected.run_id
         AND checkpoint.relation_ordinal=expected.relation_ordinal
         AND checkpoint.checkpoint_key=expected.checkpoint_key
         AND checkpoint.state='complete'
        WHERE expected.operation_id=p_operation_id AND expected.run_id=p_run_id
          AND checkpoint.operation_id IS NULL
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Notifications V1 projection checkpoint is incomplete';
    END IF;
    SELECT checkpoint.final_lsn INTO STRICT v_final_lsn
    FROM platform.cutover_copy_checkpoints AS checkpoint
    WHERE checkpoint.operation_id=p_operation_id AND checkpoint.run_id=p_run_id
    LIMIT 1;
    IF EXISTS (
        SELECT 1 FROM (
            SELECT stage.source_identity
            FROM notifications_v1_adoption_stage AS stage
            WHERE stage.operation_id=p_operation_id AND stage.run_id=p_run_id
              AND NOT stage.removed GROUP BY stage.source_identity
        ) AS source
        WHERE (SELECT pg_catalog.count(*)
               FROM notifications_v1_adoption_stage AS primary_row
               WHERE primary_row.operation_id=p_operation_id
                 AND primary_row.run_id=p_run_id AND NOT primary_row.removed
                 AND primary_row.source_identity=source.source_identity
                 AND primary_row.target_kind='message')<>1
           OR (SELECT pg_catalog.count(*)
               FROM notifications_v1_adoption_stage AS effect_row
               WHERE effect_row.operation_id=p_operation_id
                 AND effect_row.run_id=p_run_id AND NOT effect_row.removed
                 AND effect_row.source_identity=source.source_identity
                 AND effect_row.target_kind='effect')<>1
           OR (SELECT pg_catalog.count(*)
               FROM notifications_v1_adoption_stage AS delivery_row
               WHERE delivery_row.operation_id=p_operation_id
                 AND delivery_row.run_id=p_run_id AND NOT delivery_row.removed
                 AND delivery_row.source_identity=source.source_identity
                 AND delivery_row.target_kind='delivery')
              <>(SELECT pg_catalog.count(*)
                 FROM notifications_v1_adoption_stage AS attempt_row
                 WHERE attempt_row.operation_id=p_operation_id
                   AND attempt_row.run_id=p_run_id AND NOT attempt_row.removed
                   AND attempt_row.source_identity=source.source_identity
                   AND attempt_row.target_kind='attempt')
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Notifications V1 source-target mapping is incomplete';
    END IF;
    SELECT (SELECT pg_catalog.count(*) FROM notifications_inbox_counters)+
        (SELECT pg_catalog.count(*) FROM notifications_messages)+
        (SELECT pg_catalog.count(*) FROM notifications_effect_receipts)+
        (SELECT pg_catalog.count(*) FROM notifications_deliveries)+
        (SELECT pg_catalog.count(*) FROM notifications_delivery_attempts)+
        (SELECT pg_catalog.count(*) FROM notifications_delivery_receipts)+
        (SELECT pg_catalog.count(*) FROM notifications_v1_cutover_receipts)
    INTO v_existing;
    IF v_existing<>0 THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Notifications V1 inactive target is not empty';
    END IF;
    v_assert := notifications_v1_adoption_queue_evidence_v1(
        p_operation_id,p_run_id,p_scope_id
    );
    v_assert := notifications_v1_adoption_legacy_evidence_v1(
        p_operation_id,p_run_id,p_scope_id
    );
    v_assert := notifications_v1_adoption_bridge_disabled_v1(p_scope_id);
    RETURN v_final_lsn;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION notifications_v1_adoption_projection_ready_v1(
    uuid,uuid,uuid
) FROM PUBLIC;
