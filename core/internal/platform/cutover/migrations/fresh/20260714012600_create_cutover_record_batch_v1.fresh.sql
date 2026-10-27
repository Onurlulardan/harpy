-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.cutover_record_batch_v1(
    p_operation_id uuid,
    p_run_id uuid,
    p_relation_ordinal smallint,
    p_checkpoint_key text,
    p_snapshot_name text,
    p_snapshot_lsn pg_lsn,
    p_received_lsn pg_lsn,
    p_commit_lsn pg_lsn,
    p_final_lsn pg_lsn,
    p_source_xid xid8,
    p_batch_ordinal integer,
    p_checkpoint_state text,
    p_last_key jsonb,
    p_batch_sha256 text,
    p_applied_bytes bigint,
    p_applied_count bigint,
    p_key_state_sha256 text,
    p_state_sha256 text
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_manifest_sha256 text;
BEGIN
    IF SESSION_USER <> 'forgecrud_adopter' THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'cutover batch receipt principal denied';
    END IF;
    IF p_applied_count < 0 OR p_applied_bytes < 1 OR p_batch_ordinal < 1 OR
       p_batch_sha256 !~ '^[0-9a-f]{64}$' OR
       p_key_state_sha256 !~ '^[0-9a-f]{64}$' OR
       p_state_sha256 !~ '^[0-9a-f]{64}$' OR
       p_checkpoint_state NOT IN ('initial_copy', 'delta', 'complete') OR
       p_snapshot_lsn > p_commit_lsn OR p_commit_lsn > p_received_lsn OR
       (p_checkpoint_state = 'complete') <> (p_final_lsn IS NOT NULL) OR
       (p_final_lsn IS NOT NULL AND p_commit_lsn < p_final_lsn) THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'cutover batch receipt input invalid';
    END IF;
    SELECT operation.lease_manifest_sha256 INTO STRICT v_manifest_sha256
    FROM platform.cutover_operations AS operation
    WHERE operation.operation_id = p_operation_id AND operation.run_id = p_run_id
    FOR UPDATE;

    INSERT INTO platform.cutover_batch_receipts (
        operation_id, run_id, relation_ordinal, checkpoint_key, batch_ordinal,
        source_xid, source_commit_lsn, batch_sha256, applied_count,
        key_state_sha256, state_sha256
    ) VALUES (
        p_operation_id, p_run_id, p_relation_ordinal, p_checkpoint_key, p_batch_ordinal,
        p_source_xid, p_commit_lsn, p_batch_sha256, p_applied_count,
        p_key_state_sha256, p_state_sha256
    );

    INSERT INTO platform.cutover_copy_checkpoints (
        operation_id, run_id, relation_ordinal, checkpoint_key, manifest_sha256,
        snapshot_name, snapshot_lsn, last_key, received_lsn, applied_lsn, final_lsn,
        applied_rows, applied_bytes, attempt, state, blocked_reason
    ) VALUES (
        p_operation_id, p_run_id, p_relation_ordinal, p_checkpoint_key, v_manifest_sha256,
        p_snapshot_name, p_snapshot_lsn, p_last_key, p_received_lsn, p_commit_lsn, p_final_lsn,
        p_applied_count, p_applied_bytes, 1, p_checkpoint_state, NULL
    ) ON CONFLICT (operation_id, run_id, relation_ordinal, checkpoint_key) DO UPDATE
    SET last_key = EXCLUDED.last_key,
        received_lsn = EXCLUDED.received_lsn,
        applied_lsn = EXCLUDED.applied_lsn,
        final_lsn = EXCLUDED.final_lsn,
        applied_rows = platform.cutover_copy_checkpoints.applied_rows + EXCLUDED.applied_rows,
        applied_bytes = platform.cutover_copy_checkpoints.applied_bytes + EXCLUDED.applied_bytes,
        attempt = platform.cutover_copy_checkpoints.attempt + 1,
        state = EXCLUDED.state,
        blocked_reason = NULL,
        updated_at = pg_catalog.clock_timestamp()
    WHERE platform.cutover_copy_checkpoints.manifest_sha256 = EXCLUDED.manifest_sha256
      AND platform.cutover_copy_checkpoints.snapshot_name = EXCLUDED.snapshot_name
      AND platform.cutover_copy_checkpoints.snapshot_lsn = EXCLUDED.snapshot_lsn
      AND platform.cutover_copy_checkpoints.applied_lsn <= EXCLUDED.applied_lsn
      AND platform.cutover_copy_checkpoints.received_lsn <= EXCLUDED.received_lsn;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '22000', MESSAGE = 'cutover checkpoint regression detected';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.cutover_record_batch_v1(
    uuid, uuid, smallint, text, text, pg_lsn, pg_lsn, pg_lsn, pg_lsn,
    xid8, integer, text, jsonb, text, bigint, bigint, text, text
) FROM PUBLIC;
