-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_v1_adoption_stage_plan_v1(
    p_operation_id uuid, p_run_id uuid, p_scope_kind text, p_scope_id uuid,
    p_relation_ordinal smallint, p_checkpoint_key text, p_commit_lsn pg_lsn,
    p_source_xid xid8, p_batch_ordinal integer, p_batch jsonb,
    p_batch_sha256 text, p_lease_manifest_sha256 text
)
RETURNS TABLE (
    applied_count bigint, key_state_sha256 text,
    state_sha256 text, plan_bytes bytea
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_change jsonb;
    v_after jsonb;
    v_chunk bytea;
    v_applied bigint;
    v_count bigint;
    v_expected integer;
    v_min integer;
    v_max integer;
    v_distinct_counts bigint;
    v_digest text;
    v_plan bytea;
    v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    v_applied := 0;
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_batch_sha256 !~ '^[0-9a-f]{64}$' OR
       p_lease_manifest_sha256 !~ '^[0-9a-f]{64}$' OR NOT EXISTS (
           SELECT 1 FROM platform.cutover_operations AS operation
           WHERE operation.operation_id = p_operation_id
             AND operation.run_id = p_run_id
             AND operation.scope_kind = p_scope_kind
             AND operation.scope_id = p_scope_id
             AND operation.lease_manifest_sha256 = p_lease_manifest_sha256
             AND operation.state IN ('copying','shadow','quiescing')
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing V1 adoption plan staging denied';
    END IF;
    FOR v_change IN
        SELECT value
        FROM pg_catalog.jsonb_array_elements(p_batch->'changes')
             WITH ORDINALITY AS item(value, position)
        ORDER BY position
    LOOP
        v_after := v_change->'after';
        v_chunk := pg_catalog.decode(v_after->>'chunk_base64', 'base64');
        INSERT INTO platform.billing_v1_adoption_plan_chunks (
            operation_id, run_id, scope_kind, scope_id,
            chunk_ordinal, chunk_count, chunk_sha256, chunk_bytes,
            source_xid, commit_lsn, batch_ordinal, change_ordinal, created_at
        ) VALUES (
            p_operation_id, p_run_id, p_scope_kind, p_scope_id,
            (v_after->>'chunk_ordinal')::integer,
            (v_after->>'chunk_count')::integer,
            pg_catalog.decode(v_after->>'chunk_sha256', 'hex'), v_chunk,
            p_source_xid, p_commit_lsn, p_batch_ordinal,
            (v_change->>'ordinal')::integer, v_now
        );
        v_applied := v_applied + 1;
        INSERT INTO platform.applied_change_receipts (
            receipt_id, operation_id, run_id, manifest_sha256,
            source_xid, source_commit_lsn, relation_ordinal, change_ordinal,
            checkpoint_key, mutation_kind, mutation_sha256, applied_at
        ) VALUES (
            pg_catalog.gen_random_uuid(), p_operation_id, p_run_id,
            p_lease_manifest_sha256, p_source_xid, p_commit_lsn,
            p_relation_ordinal,
            (((p_batch_ordinal - 1)::bigint * 1000) +
                (v_change->>'ordinal')::integer)::integer,
            p_checkpoint_key, 'insert',
            pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
                p_batch_sha256 || ':' || (v_change->>'ordinal'), 'UTF8'
            )), 'hex'), v_now
        );
    END LOOP;
    SELECT pg_catalog.count(*), pg_catalog.min(chunk.chunk_count),
        pg_catalog.min(chunk.chunk_ordinal), pg_catalog.max(chunk.chunk_ordinal),
        pg_catalog.count(DISTINCT chunk.chunk_count)
    INTO v_count, v_expected, v_min, v_max, v_distinct_counts
    FROM platform.billing_v1_adoption_plan_chunks AS chunk
    WHERE chunk.operation_id = p_operation_id AND chunk.run_id = p_run_id
      AND chunk.scope_kind = p_scope_kind AND chunk.scope_id = p_scope_id;
    IF v_count > 0 AND v_distinct_counts IS DISTINCT FROM 1::bigint THEN
        RAISE EXCEPTION USING ERRCODE = '22000',
            MESSAGE = 'billing V1 adoption chunk cardinality differs';
    END IF;
    SELECT pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
        COALESCE(pg_catalog.string_agg(
            chunk.chunk_ordinal::text || ':' ||
            pg_catalog.encode(chunk.chunk_sha256, 'hex'),
            ',' ORDER BY chunk.chunk_ordinal
        ), ''), 'UTF8'
    )), 'hex')
    INTO v_digest
    FROM platform.billing_v1_adoption_plan_chunks AS chunk
    WHERE chunk.operation_id = p_operation_id AND chunk.run_id = p_run_id;
    IF v_count = v_expected AND v_min = 1 AND v_max = v_expected THEN
        SELECT pg_catalog.decode(pg_catalog.string_agg(
            pg_catalog.encode(chunk.chunk_bytes, 'hex'),
            '' ORDER BY chunk.chunk_ordinal
        ), 'hex')
        INTO v_plan
        FROM platform.billing_v1_adoption_plan_chunks AS chunk
        WHERE chunk.operation_id = p_operation_id AND chunk.run_id = p_run_id;
    END IF;
    RETURN QUERY SELECT v_applied, v_digest, v_digest, v_plan;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_v1_adoption_stage_plan_v1(
    uuid, uuid, text, uuid, smallint, text, pg_lsn, xid8,
    integer, jsonb, text, text
) FROM PUBLIC;
