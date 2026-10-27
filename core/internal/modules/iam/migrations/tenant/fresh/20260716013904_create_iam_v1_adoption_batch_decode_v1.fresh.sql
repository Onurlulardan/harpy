-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_v1_adoption_batch_decode_v1(
    p_change_batch_bytes bytea,
    p_batch_sha256 text,
    p_batch_bytes bigint
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_batch jsonb;
    v_actual_bytes bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM V1 adoption batch principal denied';
    END IF;
    v_actual_bytes := pg_catalog.octet_length(p_change_batch_bytes);
    IF p_batch_bytes IS DISTINCT FROM v_actual_bytes OR
       v_actual_bytes NOT BETWEEN 2 AND 8388608 OR
       p_batch_sha256 !~ '^[0-9a-f]{64}$' OR
       pg_catalog.encode(pg_catalog.sha256(p_change_batch_bytes), 'hex')
           IS DISTINCT FROM p_batch_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM V1 adoption batch digest is invalid';
    END IF;
    v_batch := pg_catalog.convert_from(p_change_batch_bytes, 'UTF8')::jsonb;
    IF pg_catalog.jsonb_typeof(v_batch) IS DISTINCT FROM 'object' OR
       (SELECT pg_catalog.array_agg(key ORDER BY key)
        FROM pg_catalog.jsonb_object_keys(v_batch) AS key)
       IS DISTINCT FROM ARRAY[
           'changes', 'checkpoint_key', 'checkpoint_state', 'last_key',
           'operation_id', 'relation_ordinal', 'run_id', 'scope_id',
           'scope_kind', 'source_position'
       ]::text[] OR
       pg_catalog.jsonb_typeof(v_batch->'source_position') IS DISTINCT FROM 'object' OR
       (SELECT pg_catalog.array_agg(key ORDER BY key)
        FROM pg_catalog.jsonb_object_keys(v_batch->'source_position') AS key)
       IS DISTINCT FROM ARRAY[
           'batch_ordinal', 'commit_lsn', 'final_lsn', 'received_lsn',
           'snapshot_lsn', 'snapshot_name', 'source_xid'
       ]::text[] OR
       pg_catalog.jsonb_typeof(v_batch->'changes') IS DISTINCT FROM 'array' OR
       pg_catalog.jsonb_array_length(v_batch->'changes') > 1000 OR
       (pg_catalog.jsonb_typeof(v_batch->'last_key') NOT IN ('object', 'null')) OR
       v_batch->'last_key' = '{}'::jsonb THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM V1 adoption batch envelope is invalid';
    END IF;
    IF EXISTS (
        SELECT 1
        FROM pg_catalog.jsonb_array_elements(v_batch->'changes')
             WITH ORDINALITY AS change(value, position)
        WHERE pg_catalog.jsonb_typeof(change.value) IS DISTINCT FROM 'object'
           OR (SELECT pg_catalog.array_agg(key ORDER BY key)
               FROM pg_catalog.jsonb_object_keys(change.value) AS key)
              IS DISTINCT FROM ARRAY['after','before','key','kind','ordinal']::text[]
           OR (change.value->>'ordinal')::integer IS DISTINCT FROM change.position::integer
           OR change.value->>'kind' NOT IN ('insert', 'update', 'delete')
           OR pg_catalog.jsonb_typeof(change.value->'key') IS DISTINCT FROM 'object'
           OR (SELECT pg_catalog.array_agg(key ORDER BY key)
               FROM pg_catalog.jsonb_object_keys(change.value->'key') AS key)
              IS DISTINCT FROM ARRAY[
                  'source_category','source_key','target_key','target_table'
              ]::text[]
           OR change.value->'key' @> '{"source_category":"","source_key":"","target_key":"","target_table":""}'::jsonb
           OR (change.value->>'kind' = 'insert' AND (
                  change.value->'before' <> 'null'::jsonb OR
                  NOT iam_v1_adoption_target_record_valid_v1(
                      change.value->'key', change.value->'after'
                  )))
           OR (change.value->>'kind' = 'update' AND (
                  NOT iam_v1_adoption_target_record_valid_v1(
                      change.value->'key', change.value->'before'
                  ) OR NOT iam_v1_adoption_target_record_valid_v1(
                      change.value->'key', change.value->'after'
                  )))
           OR (change.value->>'kind' = 'delete' AND (
                  change.value->'after' <> 'null'::jsonb OR
                  NOT iam_v1_adoption_target_record_valid_v1(
                      change.value->'key', change.value->'before'
                  )))
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM V1 adoption batch change is invalid';
    END IF;
    RETURN v_batch;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_v1_adoption_batch_decode_v1(
    bytea, text, bigint
) FROM PUBLIC;

