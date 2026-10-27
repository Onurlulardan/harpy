-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION notifications_v1_adoption_apply_and_record_v1(
    p_operation_id uuid,p_run_id uuid,p_scope_kind text,p_scope_id uuid,
    p_relation_ordinal smallint,p_checkpoint_key text,p_snapshot_name text,
    p_snapshot_lsn pg_lsn,p_received_lsn pg_lsn,p_commit_lsn pg_lsn,
    p_final_lsn pg_lsn,p_source_xid xid8,p_batch_ordinal integer,
    p_checkpoint_state text,p_last_key jsonb,p_change_batch_bytes bytea,
    p_batch_sha256 text,p_batch_bytes bigint,p_lease_manifest_sha256 text
)
RETURNS TABLE (applied_count bigint,key_state_sha256 text,state_sha256 text)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_batch jsonb;
    v_contract_id text;
    v_procedure_oid bigint;
    v_procedure_signature text;
    v_expected_signature text;
    v_already_applied boolean;
    v_applied_count bigint;
    v_key_sha text;
    v_state_sha text;
    v_projected_sha text;
    v_asserted boolean;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Notifications V1 adoption composite principal denied';
    END IF;
    v_batch := notifications_v1_adoption_batch_decode_v1(
        p_change_batch_bytes,p_batch_sha256,p_batch_bytes
    );
    v_asserted := notifications_v1_adoption_batch_assert_v1(
        v_batch,p_operation_id,p_run_id,p_scope_kind,p_scope_id,
        p_relation_ordinal,p_checkpoint_key,p_snapshot_name,
        p_snapshot_lsn,p_received_lsn,p_commit_lsn,p_final_lsn,
        p_source_xid,p_batch_ordinal,p_checkpoint_state,p_last_key
    );
    IF NOT v_asserted THEN
        RAISE EXCEPTION USING ERRCODE='22000',
            MESSAGE='Notifications V1 adoption batch assertion failed';
    END IF;
    SELECT prepared.procedure_contract_id,prepared.procedure_oid,
        prepared.procedure_signature,prepared.already_applied,
        prepared.applied_count,prepared.key_state_sha256,prepared.state_sha256
    INTO STRICT v_contract_id,v_procedure_oid,v_procedure_signature,
        v_already_applied,v_applied_count,v_key_sha,v_state_sha
    FROM platform.cutover_prepare_batch_v1(
        p_operation_id,p_run_id,p_scope_kind,p_scope_id,
        p_relation_ordinal,p_checkpoint_key,p_batch_ordinal,
        p_commit_lsn,p_source_xid,p_batch_sha256,p_lease_manifest_sha256
    ) AS prepared;
    v_expected_signature := pg_catalog.format(
        '%I.notifications_v1_adoption_apply_v1(uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)',
        pg_catalog.current_schema()
    );
    IF v_contract_id IS DISTINCT FROM 'notifications.v1_adoption.v1' OR
       v_procedure_oid IS DISTINCT FROM
           'notifications_v1_adoption_apply_v1(uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)'::regprocedure::oid OR
       v_procedure_signature IS DISTINCT FROM v_expected_signature THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Notifications V1 adoption procedure binding differs';
    END IF;
    IF v_already_applied THEN
        RETURN QUERY SELECT v_applied_count,v_key_sha,v_state_sha;
        RETURN;
    END IF;
    SELECT result.applied_count,result.key_state_sha256,result.state_sha256
    INTO STRICT v_applied_count,v_key_sha,v_state_sha
    FROM notifications_v1_adoption_apply_v1(
        p_operation_id,p_run_id,p_scope_kind,p_scope_id,
        p_relation_ordinal,p_checkpoint_key,p_snapshot_lsn,p_received_lsn,
        p_commit_lsn,p_source_xid,p_batch_ordinal,v_batch,p_batch_sha256
    ) AS result;
    IF v_applied_count<0 OR v_key_sha !~ '^[0-9a-f]{64}$' OR
       v_state_sha !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE='22000',
            MESSAGE='Notifications V1 adoption owner result is invalid';
    END IF;
    PERFORM platform.cutover_record_batch_v1(
        p_operation_id,p_run_id,p_relation_ordinal,p_checkpoint_key,
        p_snapshot_name,p_snapshot_lsn,p_received_lsn,p_commit_lsn,
        p_final_lsn,p_source_xid,p_batch_ordinal,p_checkpoint_state,
        p_last_key,p_batch_sha256,p_batch_bytes,v_applied_count,
        v_key_sha,v_state_sha
    );
    IF p_checkpoint_state='complete' AND NOT EXISTS (
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
        v_projected_sha := notifications_v1_adoption_project_v1(
            p_operation_id,p_run_id,p_scope_id
        );
        IF v_projected_sha IS DISTINCT FROM v_state_sha THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='Notifications V1 adoption projected digest differs';
        END IF;
    END IF;
    RETURN QUERY SELECT v_applied_count,v_key_sha,v_state_sha;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION notifications_v1_adoption_apply_and_record_v1(
    uuid,uuid,text,uuid,smallint,text,text,pg_lsn,pg_lsn,pg_lsn,
    pg_lsn,xid8,integer,text,jsonb,bytea,text,bigint,text
) FROM PUBLIC;
