-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_v1_adoption_global_apply_v1(
    p_operation_id uuid, p_run_id uuid, p_scope_kind text, p_scope_id uuid,
    p_relation_ordinal smallint, p_checkpoint_key text,
    p_snapshot_lsn pg_lsn, p_received_lsn pg_lsn, p_commit_lsn pg_lsn,
    p_source_xid xid8, p_batch_ordinal integer, p_batch jsonb,
    p_batch_sha256 text
)
RETURNS TABLE (applied_count bigint, key_state_sha256 text, state_sha256 text)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_applied bigint;
    v_key_sha text;
    v_state_sha text;
    v_plan bytea;
    v_source_sha text;
    v_lease_manifest text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_kind IS DISTINCT FROM 'global' OR
       p_checkpoint_key IS DISTINCT FROM 'catalog' OR
       p_relation_ordinal IS DISTINCT FROM 1::smallint THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing global V1 adoption owner apply denied';
    END IF;
    SELECT operation.lease_manifest_sha256 INTO STRICT v_lease_manifest
    FROM platform.cutover_operations AS operation
    WHERE operation.operation_id = p_operation_id AND operation.run_id = p_run_id
      AND operation.scope_kind = 'global' AND operation.scope_id = p_scope_id
      AND operation.group_id = 'billing' AND operation.domain = 'billing';
    SELECT staged.applied_count, staged.key_state_sha256,
        staged.state_sha256, staged.plan_bytes
    INTO STRICT v_applied, v_key_sha, v_state_sha, v_plan
    FROM platform.billing_v1_adoption_stage_plan_v1(
        p_operation_id, p_run_id, p_scope_kind, p_scope_id,
        p_relation_ordinal, p_checkpoint_key, p_commit_lsn, p_source_xid,
        p_batch_ordinal, p_batch, p_batch_sha256, v_lease_manifest
    ) AS staged;
    IF v_plan IS NOT NULL THEN
        v_source_sha := platform.billing_v1_adoption_register_apply_v1(
            p_operation_id, p_run_id, p_scope_kind, p_scope_id, v_plan,
            p_batch->'last_key', p_batch_sha256, p_batch_ordinal,
            v_lease_manifest
        );
    END IF;
    IF p_batch->>'checkpoint_state' = 'complete' THEN
        IF v_plan IS NULL OR v_source_sha IS NULL THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'billing global V1 adoption plan is incomplete';
        END IF;
        v_state_sha := platform.billing_v1_adoption_verify_rollback_v1(
            p_operation_id, p_run_id, v_source_sha
        );
    END IF;
    RETURN QUERY SELECT v_applied, v_key_sha, v_state_sha;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_v1_adoption_global_apply_v1(
    uuid, uuid, text, uuid, smallint, text, pg_lsn, pg_lsn, pg_lsn,
    xid8, integer, jsonb, text
) FROM PUBLIC;
