-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.cutover_probe_apply_and_record_v1(
    p_operation_id uuid, p_run_id uuid, p_scope_kind text, p_scope_id uuid,
    p_relation_ordinal smallint, p_checkpoint_key text, p_snapshot_name text,
    p_snapshot_lsn pg_lsn, p_received_lsn pg_lsn, p_commit_lsn pg_lsn,
    p_final_lsn pg_lsn, p_source_xid xid8, p_batch_ordinal integer,
    p_checkpoint_state text, p_last_key jsonb, p_change_batch_bytes bytea,
    p_batch_sha256 text, p_batch_bytes bigint, p_lease_manifest_sha256 text
)
RETURNS TABLE (applied_count bigint, key_state_sha256 text, state_sha256 text)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_operation platform.cutover_operations%ROWTYPE;
    v_procedure platform.cutover_operation_procedures%ROWTYPE;
    v_receipt platform.cutover_batch_receipts%ROWTYPE;
    v_applied_count bigint;
    v_key_state_sha256 text;
    v_state_sha256 text;
    v_change_batch jsonb;
    v_actual_batch_bytes bigint;
    v_change_count integer;
BEGIN
    IF SESSION_USER <> 'forgecrud_adopter' THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'cutover composite principal denied';
    END IF;
    v_actual_batch_bytes := pg_catalog.octet_length(p_change_batch_bytes);
    IF p_batch_bytes IS DISTINCT FROM v_actual_batch_bytes OR
       v_actual_batch_bytes < 1 OR v_actual_batch_bytes > 8388608 THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'cutover change batch size invalid';
    END IF;
    IF pg_catalog.encode(pg_catalog.sha256(p_change_batch_bytes), 'hex') IS DISTINCT FROM p_batch_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '22000', MESSAGE = 'cutover change batch digest differs';
    END IF;
    v_change_batch := pg_catalog.convert_from(p_change_batch_bytes, 'UTF8')::jsonb;
    v_change_count := pg_catalog.jsonb_array_length(v_change_batch->'changes');
    IF pg_catalog.octet_length(v_change_batch::text) > 8388608 OR
       (SELECT pg_catalog.array_agg(key ORDER BY key) FROM pg_catalog.jsonb_object_keys(v_change_batch) AS key)
         IS DISTINCT FROM ARRAY['changes','checkpoint_key','checkpoint_state','last_key','operation_id',
           'relation_ordinal','run_id','scope_id','scope_kind','source_position']::text[] OR
       pg_catalog.jsonb_typeof(v_change_batch->'source_position') IS DISTINCT FROM 'object' OR
       (SELECT pg_catalog.array_agg(key ORDER BY key)
          FROM pg_catalog.jsonb_object_keys(v_change_batch->'source_position') AS key)
         IS DISTINCT FROM ARRAY['batch_ordinal','commit_lsn','final_lsn','received_lsn',
           'snapshot_lsn','snapshot_name','source_xid']::text[] OR
       (pg_catalog.jsonb_typeof(v_change_batch->'last_key') IS DISTINCT FROM 'object' AND
         pg_catalog.jsonb_typeof(v_change_batch->'last_key') IS DISTINCT FROM 'null') OR
       v_change_batch->'last_key' = '{}'::jsonb OR
        pg_catalog.jsonb_typeof(v_change_batch->'changes') IS DISTINCT FROM 'array' OR
        v_change_count > 1000 OR p_batch_ordinal NOT BETWEEN 1 AND 2147483 OR
        (v_change_count = 0 AND (p_checkpoint_state IS DISTINCT FROM 'complete' OR
          p_final_lsn IS NULL OR p_commit_lsn IS DISTINCT FROM p_final_lsn)) OR
       (v_change_batch->>'operation_id')::uuid IS DISTINCT FROM p_operation_id OR
       (v_change_batch->>'run_id')::uuid IS DISTINCT FROM p_run_id OR
       v_change_batch->>'scope_kind' IS DISTINCT FROM p_scope_kind OR
       (v_change_batch->>'scope_id')::uuid IS DISTINCT FROM p_scope_id OR
       (v_change_batch->>'relation_ordinal')::smallint IS DISTINCT FROM p_relation_ordinal OR
       v_change_batch->>'checkpoint_key' IS DISTINCT FROM p_checkpoint_key OR
       v_change_batch->>'checkpoint_state' IS DISTINCT FROM p_checkpoint_state OR
       v_change_batch->'last_key' IS DISTINCT FROM p_last_key OR
       v_change_batch->'source_position'->>'snapshot_name' IS DISTINCT FROM p_snapshot_name OR
       (v_change_batch->'source_position'->>'snapshot_lsn')::pg_lsn IS DISTINCT FROM p_snapshot_lsn OR
       (v_change_batch->'source_position'->>'received_lsn')::pg_lsn IS DISTINCT FROM p_received_lsn OR
       (v_change_batch->'source_position'->>'commit_lsn')::pg_lsn IS DISTINCT FROM p_commit_lsn OR
       (v_change_batch->'source_position'->>'final_lsn')::pg_lsn IS DISTINCT FROM p_final_lsn OR
       (v_change_batch->'source_position'->>'source_xid')::xid8 IS DISTINCT FROM p_source_xid OR
       (v_change_batch->'source_position'->>'batch_ordinal')::integer IS DISTINCT FROM p_batch_ordinal THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'cutover change batch evidence differs';
    END IF;
    IF EXISTS (
        SELECT 1 FROM pg_catalog.jsonb_array_elements(v_change_batch->'changes')
          WITH ORDINALITY AS change(value, position)
        WHERE pg_catalog.jsonb_typeof(change.value) IS DISTINCT FROM 'object'
           OR (SELECT pg_catalog.array_agg(key ORDER BY key)
                 FROM pg_catalog.jsonb_object_keys(change.value) AS key)
                IS DISTINCT FROM ARRAY['after','before','key','kind','ordinal']::text[]
           OR (change.value->>'ordinal')::integer IS DISTINCT FROM change.position::integer
           OR pg_catalog.jsonb_typeof(change.value->'kind') IS DISTINCT FROM 'string'
           OR change.value->>'kind' NOT IN ('insert','update','delete')
           OR pg_catalog.jsonb_typeof(change.value->'key') IS DISTINCT FROM 'object'
           OR change.value->'key' = '{}'::jsonb
           OR (change.value->>'kind' = 'insert' AND (change.value->'before' <> 'null'::jsonb OR
                pg_catalog.jsonb_typeof(change.value->'after') IS DISTINCT FROM 'object' OR change.value->'after' = '{}'::jsonb))
           OR (change.value->>'kind' = 'update' AND (pg_catalog.jsonb_typeof(change.value->'before') IS DISTINCT FROM 'object' OR
                change.value->'before' = '{}'::jsonb OR pg_catalog.jsonb_typeof(change.value->'after') IS DISTINCT FROM 'object' OR change.value->'after' = '{}'::jsonb))
           OR (change.value->>'kind' = 'delete' AND (pg_catalog.jsonb_typeof(change.value->'before') IS DISTINCT FROM 'object' OR
                change.value->'before' = '{}'::jsonb OR change.value->'after' <> 'null'::jsonb))
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'cutover change batch shape differs';
    END IF;
    SELECT operation.* INTO STRICT v_operation
    FROM platform.cutover_operations AS operation
    WHERE operation.operation_id = p_operation_id AND operation.run_id = p_run_id
    FOR UPDATE;
    IF v_operation.lease_manifest_sha256 IS DISTINCT FROM p_lease_manifest_sha256 OR
       v_operation.scope_kind IS DISTINCT FROM p_scope_kind OR
       v_operation.scope_id IS DISTINCT FROM p_scope_id OR
       v_operation.state NOT IN ('copying', 'shadow', 'quiescing') OR
       v_operation.lease_expires_at <= pg_catalog.clock_timestamp() OR
       NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles AS role
           WHERE role.rolname = 'forgecrud_adopter' AND NOT role.rolcanlogin
             AND NOT role.rolinherit AND NOT role.rolsuper AND NOT role.rolcreatedb
             AND NOT role.rolcreaterole AND NOT role.rolreplication AND NOT role.rolbypassrls
             AND role.rolconnlimit = 0
              AND role.rolvaliduntil IS NOT DISTINCT FROM v_operation.lease_expires_at
              AND role.rolvaliduntil > pg_catalog.clock_timestamp()
              AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_auth_members AS membership
                  WHERE membership.roleid = role.oid OR membership.member = role.oid)) OR
       NOT EXISTS (SELECT 1 FROM pg_catalog.pg_roles AS role
           WHERE role.rolname = 'forgecrud_cutover_decoder' AND NOT role.rolcanlogin
             AND NOT role.rolinherit AND NOT role.rolsuper AND NOT role.rolcreatedb
             AND NOT role.rolcreaterole AND NOT role.rolreplication AND NOT role.rolbypassrls
             AND role.rolconnlimit = 0
             AND role.rolvaliduntil IS NOT DISTINCT FROM v_operation.lease_expires_at
             AND role.rolvaliduntil > pg_catalog.clock_timestamp()
             AND NOT EXISTS (SELECT 1 FROM pg_catalog.pg_auth_members AS membership
                 WHERE membership.roleid = role.oid OR membership.member = role.oid)) THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'cutover operation is not apply-ready';
    END IF;
    SELECT procedure_row.* INTO STRICT v_procedure
    FROM platform.cutover_operation_procedures AS procedure_row
    WHERE procedure_row.operation_id = p_operation_id
      AND procedure_row.run_id = p_run_id
      AND procedure_row.relation_ordinal = p_relation_ordinal
      AND procedure_row.procedure_kind = 'apply';
    IF v_procedure.procedure_contract_id <> 'platform.cutover_probe.v1' OR
       v_procedure.procedure_oid <> 'platform.cutover_probe_apply_v1(uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)'::regprocedure::oid OR
       v_procedure.procedure_signature <> 'platform.cutover_probe_apply_v1(uuid,uuid,text,uuid,smallint,text,pg_lsn,pg_lsn,pg_lsn,xid8,integer,jsonb,text)' THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'cutover composite procedure binding differs';
    END IF;
    SELECT receipt.* INTO v_receipt
    FROM platform.cutover_batch_receipts AS receipt
    WHERE receipt.operation_id = p_operation_id AND receipt.run_id = p_run_id
      AND receipt.relation_ordinal = p_relation_ordinal
      AND receipt.checkpoint_key = p_checkpoint_key
      AND receipt.batch_ordinal = p_batch_ordinal;
    IF FOUND THEN
        IF v_receipt.batch_sha256 <> p_batch_sha256 OR
           v_receipt.source_commit_lsn <> p_commit_lsn OR
           v_receipt.source_xid <> p_source_xid THEN
            RAISE EXCEPTION USING ERRCODE = '22000', MESSAGE = 'cutover composite idempotency tamper detected';
        END IF;
        RETURN QUERY SELECT v_receipt.applied_count,
            v_receipt.key_state_sha256, v_receipt.state_sha256;
        RETURN;
    END IF;
    SELECT result.applied_count, result.key_state_sha256, result.state_sha256
    INTO STRICT v_applied_count, v_key_state_sha256, v_state_sha256
    FROM platform.cutover_probe_apply_v1(
        p_operation_id, p_run_id, p_scope_kind, p_scope_id, p_relation_ordinal,
        p_checkpoint_key, p_snapshot_lsn, p_received_lsn, p_commit_lsn,
        p_source_xid, p_batch_ordinal, v_change_batch, p_batch_sha256
    ) AS result;
    IF v_applied_count < 0 OR v_key_state_sha256 !~ '^[0-9a-f]{64}$' OR
       v_state_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '22000', MESSAGE = 'cutover owner result invalid';
    END IF;
    PERFORM platform.cutover_record_batch_v1(
        p_operation_id, p_run_id, p_relation_ordinal, p_checkpoint_key,
        p_snapshot_name, p_snapshot_lsn, p_received_lsn, p_commit_lsn,
        p_final_lsn, p_source_xid, p_batch_ordinal, p_checkpoint_state,
        p_last_key, p_batch_sha256, p_batch_bytes, v_applied_count,
        v_key_state_sha256, v_state_sha256
    );
    RETURN QUERY SELECT v_applied_count, v_key_state_sha256, v_state_sha256;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.cutover_probe_apply_and_record_v1(
    uuid, uuid, text, uuid, smallint, text, text, pg_lsn, pg_lsn, pg_lsn,
    pg_lsn, xid8, integer, text, jsonb, bytea, text, bigint, text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.cutover_probe_apply_and_record_v1(
    uuid, uuid, text, uuid, smallint, text, text, pg_lsn, pg_lsn, pg_lsn,
    pg_lsn, xid8, integer, text, jsonb, bytea, text, bigint, text
) TO forgecrud_adopter;
