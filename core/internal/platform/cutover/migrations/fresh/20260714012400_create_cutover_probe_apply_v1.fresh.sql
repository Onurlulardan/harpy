-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.cutover_probe_apply_v1(
    p_operation_id uuid,
    p_run_id uuid,
    p_scope_kind text,
    p_scope_id uuid,
    p_relation_ordinal smallint,
    p_checkpoint_key text,
    p_snapshot_lsn pg_lsn,
    p_received_lsn pg_lsn,
    p_commit_lsn pg_lsn,
    p_source_xid xid8,
    p_batch_ordinal integer,
    p_change_batch jsonb,
    p_batch_sha256 text
)
RETURNS TABLE (applied_count bigint, key_state_sha256 text, state_sha256 text)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_existing text;
    v_count bigint;
    v_manifest_sha256 text;
BEGIN
    IF SESSION_USER <> 'harpy_adopter' THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'cutover probe apply principal denied';
    END IF;
    IF p_operation_id IS NULL OR p_run_id IS NULL OR p_scope_kind IS NULL OR
       p_scope_id IS NULL OR p_relation_ordinal IS NULL OR p_checkpoint_key IS NULL OR
       p_scope_kind NOT IN ('global', 'tenant') OR p_received_lsn < p_commit_lsn OR
       p_snapshot_lsn > p_commit_lsn OR p_batch_ordinal NOT BETWEEN 1 AND 2147483 OR
       p_batch_sha256 !~ '^[0-9a-f]{64}$' OR
       pg_catalog.octet_length(p_change_batch::text) > 8388608 OR
       pg_catalog.jsonb_typeof(p_change_batch->'changes') IS DISTINCT FROM 'array' OR
       pg_catalog.jsonb_array_length(p_change_batch->'changes') > 1000 OR
       (pg_catalog.jsonb_array_length(p_change_batch->'changes') = 0 AND
         (p_change_batch->>'checkpoint_state' IS DISTINCT FROM 'complete' OR
          p_change_batch->'source_position'->>'final_lsn' IS NULL OR
          (p_change_batch->'source_position'->>'final_lsn')::pg_lsn IS DISTINCT FROM p_commit_lsn)) OR
       EXISTS (
           SELECT 1
           FROM pg_catalog.jsonb_array_elements(p_change_batch->'changes') AS change(value)
           WHERE pg_catalog.jsonb_typeof(change.value) IS DISTINCT FROM 'object'
              OR pg_catalog.jsonb_typeof(change.value->'kind') IS DISTINCT FROM 'string'
              OR change.value->>'kind' NOT IN ('insert','update','delete')
       ) OR
       p_change_batch->>'operation_id' IS DISTINCT FROM p_operation_id::text OR
       p_change_batch->>'run_id' IS DISTINCT FROM p_run_id::text OR
       p_change_batch->>'scope_kind' IS DISTINCT FROM p_scope_kind OR
       p_change_batch->>'scope_id' IS DISTINCT FROM p_scope_id::text OR
       (p_change_batch->>'relation_ordinal')::integer IS DISTINCT FROM p_relation_ordinal OR
       p_change_batch->>'checkpoint_key' IS DISTINCT FROM p_checkpoint_key THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'cutover probe apply binding differs';
    END IF;

    INSERT INTO platform.cutover_contract_probe_rows (
        operation_id, run_id, relation_ordinal, checkpoint_key,
        batch_ordinal, batch_sha256, canonical_payload
    ) VALUES (
        p_operation_id, p_run_id, p_relation_ordinal, p_checkpoint_key,
        p_batch_ordinal, p_batch_sha256, p_change_batch
    ) ON CONFLICT DO NOTHING;

    SELECT row.batch_sha256 INTO v_existing
    FROM platform.cutover_contract_probe_rows AS row
    WHERE row.operation_id = p_operation_id AND row.run_id = p_run_id
      AND row.relation_ordinal = p_relation_ordinal
      AND row.checkpoint_key = p_checkpoint_key AND row.batch_ordinal = p_batch_ordinal;
    IF v_existing IS DISTINCT FROM p_batch_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '22000', MESSAGE = 'cutover probe batch tamper detected';
    END IF;

    v_count := pg_catalog.jsonb_array_length(p_change_batch->'changes');
    SELECT operation.lease_manifest_sha256 INTO STRICT v_manifest_sha256
    FROM platform.cutover_operations AS operation
    WHERE operation.operation_id = p_operation_id AND operation.run_id = p_run_id;
    INSERT INTO platform.applied_change_receipts (
        receipt_id, operation_id, run_id, manifest_sha256, source_xid,
        source_commit_lsn, relation_ordinal, change_ordinal, checkpoint_key,
        mutation_kind, mutation_sha256, applied_at
    ) SELECT pg_catalog.gen_random_uuid(), p_operation_id, p_run_id, v_manifest_sha256,
        p_source_xid, p_commit_lsn, p_relation_ordinal,
        (((p_batch_ordinal - 1)::bigint * 1000) + change.position)::integer,
        p_checkpoint_key, change.value->>'kind',
        pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
            p_batch_sha256 || ':' || change.position::text, 'UTF8')), 'hex'),
        pg_catalog.clock_timestamp()
    FROM pg_catalog.jsonb_array_elements(p_change_batch->'changes')
        WITH ORDINALITY AS change(value, position)
    ;
    RETURN QUERY SELECT
        v_count,
        pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
            COALESCE((p_change_batch->'last_key')::text, 'null'), 'UTF8'
        )), 'hex'),
        pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
            p_change_batch::text || ':' || v_count::text, 'UTF8'
        )), 'hex');
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.cutover_probe_apply_v1(
    uuid, uuid, text, uuid, smallint, text, pg_lsn, pg_lsn, pg_lsn,
    xid8, integer, jsonb, text
) FROM PUBLIC;
