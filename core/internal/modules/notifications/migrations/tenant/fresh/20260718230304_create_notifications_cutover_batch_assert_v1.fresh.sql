-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION notifications_v1_adoption_batch_assert_v1(
    p_batch jsonb,p_operation_id uuid,p_run_id uuid,
    p_scope_kind text,p_scope_id uuid,p_relation_ordinal smallint,
    p_checkpoint_key text,p_snapshot_name text,p_snapshot_lsn pg_lsn,
    p_received_lsn pg_lsn,p_commit_lsn pg_lsn,p_final_lsn pg_lsn,
    p_source_xid xid8,p_batch_ordinal integer,p_checkpoint_state text,p_last_key jsonb
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_count integer;
BEGIN
    v_count := pg_catalog.jsonb_array_length(p_batch->'changes');
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Notifications V1 adoption evidence principal denied';
    END IF;
    IF p_scope_kind IS DISTINCT FROM 'tenant' OR p_scope_id IS NULL OR
       p_relation_ordinal NOT BETWEEN 1 AND 32767 OR
       p_checkpoint_key IS DISTINCT FROM 'notification' OR
       p_batch_ordinal NOT BETWEEN 1 AND 2147483 OR
       p_checkpoint_state NOT IN ('initial_copy','delta','complete') OR
       (p_checkpoint_state='complete') IS DISTINCT FROM (p_final_lsn IS NOT NULL) OR
       (v_count=0 AND (
           p_checkpoint_state<>'complete' OR p_final_lsn IS DISTINCT FROM p_commit_lsn
       )) OR NOT EXISTS (
           SELECT 1 FROM tenancy_schema_identity AS identity
           WHERE identity.singleton AND identity.tenant_id=p_scope_id
             AND identity.schema_name=pg_catalog.current_schema()
             AND identity.origin='v1_adopted'
       ) OR NOT EXISTS (
           SELECT 1 FROM platform.cutover_operations AS operation
           WHERE operation.operation_id=p_operation_id AND operation.run_id=p_run_id
             AND operation.scope_kind=p_scope_kind AND operation.scope_id=p_scope_id
             AND operation.group_id='notifications' AND operation.domain='notifications'
             AND operation.state IN ('planned','copying','shadow','quiescing')
       ) OR NOT EXISTS (
           SELECT 1 FROM platform.cutover_operation_checkpoint_keys AS expected
           WHERE expected.operation_id=p_operation_id AND expected.run_id=p_run_id
             AND expected.relation_ordinal=p_relation_ordinal
             AND expected.checkpoint_key=p_checkpoint_key
       ) OR (p_batch->>'operation_id')::uuid IS DISTINCT FROM p_operation_id OR
       (p_batch->>'run_id')::uuid IS DISTINCT FROM p_run_id OR
       p_batch->>'scope_kind' IS DISTINCT FROM p_scope_kind OR
       (p_batch->>'scope_id')::uuid IS DISTINCT FROM p_scope_id OR
       (p_batch->>'relation_ordinal')::smallint IS DISTINCT FROM p_relation_ordinal OR
       p_batch->>'checkpoint_key' IS DISTINCT FROM p_checkpoint_key OR
       p_batch->>'checkpoint_state' IS DISTINCT FROM p_checkpoint_state OR
       p_batch->'last_key' IS DISTINCT FROM p_last_key OR
       p_batch->'source_position'->>'snapshot_name' IS DISTINCT FROM p_snapshot_name OR
       (p_batch->'source_position'->>'snapshot_lsn')::pg_lsn
           IS DISTINCT FROM p_snapshot_lsn OR
       (p_batch->'source_position'->>'received_lsn')::pg_lsn
           IS DISTINCT FROM p_received_lsn OR
       (p_batch->'source_position'->>'commit_lsn')::pg_lsn
           IS DISTINCT FROM p_commit_lsn OR
       (p_batch->'source_position'->>'final_lsn')::pg_lsn
           IS DISTINCT FROM p_final_lsn OR
       (p_batch->'source_position'->>'source_xid')::xid8 IS DISTINCT FROM p_source_xid OR
       (p_batch->'source_position'->>'batch_ordinal')::integer
           IS DISTINCT FROM p_batch_ordinal OR EXISTS (
           SELECT 1 FROM pg_catalog.jsonb_array_elements(p_batch->'changes') AS change
           WHERE COALESCE(change->'after'->>'source_kind',
                          change->'before'->>'source_kind')<>'notification'
       ) OR EXISTS (
           SELECT 1 FROM pg_catalog.jsonb_array_elements(p_batch->'changes') AS change
           GROUP BY COALESCE(change->'after'->>'source_identity',
                             change->'before'->>'source_identity')
           HAVING pg_catalog.count(DISTINCT COALESCE(
               change->'after'->>'source_sha256',change->'before'->>'source_sha256'
           ))<>1
       ) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Notifications V1 adoption batch evidence differs';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION notifications_v1_adoption_batch_assert_v1(
    jsonb,uuid,uuid,text,uuid,smallint,text,text,pg_lsn,pg_lsn,
    pg_lsn,pg_lsn,xid8,integer,text,jsonb
) FROM PUBLIC;
