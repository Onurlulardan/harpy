-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_v1_adoption_batch_assert_v1(
    p_batch jsonb, p_operation_id uuid, p_run_id uuid,
    p_scope_kind text, p_scope_id uuid, p_relation_ordinal smallint,
    p_checkpoint_key text, p_snapshot_name text, p_snapshot_lsn pg_lsn,
    p_received_lsn pg_lsn, p_commit_lsn pg_lsn, p_final_lsn pg_lsn,
    p_source_xid xid8, p_batch_ordinal integer, p_checkpoint_state text,
    p_last_key jsonb, p_expected_scope text, p_expected_checkpoint text
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_change_count integer;
    v_group text;
BEGIN
    v_change_count := pg_catalog.jsonb_array_length(p_batch->'changes');
    v_group := CASE p_expected_scope
        WHEN 'global' THEN 'billing' ELSE 'billing_forms' END;
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing V1 adoption evidence principal denied';
    END IF;
    IF p_expected_scope NOT IN ('global', 'tenant') OR
       p_scope_kind IS DISTINCT FROM p_expected_scope OR
       p_checkpoint_key IS DISTINCT FROM p_expected_checkpoint OR
       p_relation_ordinal IS DISTINCT FROM 1::smallint OR
       p_scope_id IS NULL OR p_batch_ordinal NOT BETWEEN 1 AND 2147483 OR
       p_checkpoint_state NOT IN ('initial_copy','delta','complete') OR
       p_snapshot_name IS NULL OR p_snapshot_name = '' OR
       p_snapshot_lsn > p_commit_lsn OR p_commit_lsn > p_received_lsn OR
       (p_checkpoint_state = 'complete') IS DISTINCT FROM
           (p_final_lsn IS NOT NULL) OR
       (p_final_lsn IS NOT NULL AND p_commit_lsn < p_final_lsn) OR
       (v_change_count = 0 AND (
           p_checkpoint_state <> 'complete' OR
           p_final_lsn IS DISTINCT FROM p_commit_lsn
       )) OR NOT EXISTS (
           SELECT 1 FROM platform.cutover_operations AS operation
           WHERE operation.operation_id = p_operation_id
             AND operation.run_id = p_run_id
             AND operation.scope_kind = p_scope_kind
             AND operation.scope_id = p_scope_id
             AND operation.group_id = v_group
             AND operation.domain = 'billing'
             AND operation.state IN ('copying','shadow','quiescing')
       ) OR NOT EXISTS (
           SELECT 1
           FROM platform.cutover_groups AS group_row
           JOIN platform.domain_cutovers AS billing_row
             ON billing_row.scope_kind = group_row.scope_kind
            AND billing_row.scope_id = group_row.scope_id
            AND billing_row.domain = 'billing'
           WHERE group_row.scope_kind = p_scope_kind
             AND group_row.scope_id = p_scope_id
             AND group_row.group_id = v_group
             AND group_row.state IN ('shadow','quiescing')
             AND billing_row.state IN ('shadow','quiescing')
       ) OR (p_expected_scope = 'global' AND NOT EXISTS (
           SELECT 1 FROM platform.environment_identity AS identity
           WHERE identity.environment_id = p_scope_id
       )) OR (p_expected_scope = 'tenant' AND (
           NOT EXISTS (
               SELECT 1 FROM platform.tenants AS tenant
               WHERE tenant.id = p_scope_id AND tenant.deleted_at IS NULL
           ) OR (SELECT pg_catalog.string_agg(member.domain, ',' ORDER BY member.ordinal)
               FROM platform.cutover_group_members AS member
               WHERE member.scope_kind = 'tenant'
                 AND member.scope_id = p_scope_id
                 AND member.group_id = 'billing_forms')
              IS DISTINCT FROM 'billing,forms'
       )) OR (p_batch->>'operation_id')::uuid IS DISTINCT FROM p_operation_id OR
       (p_batch->>'run_id')::uuid IS DISTINCT FROM p_run_id OR
       p_batch->>'scope_kind' IS DISTINCT FROM p_scope_kind OR
       (p_batch->>'scope_id')::uuid IS DISTINCT FROM p_scope_id OR
       (p_batch->>'relation_ordinal')::smallint IS DISTINCT FROM
           p_relation_ordinal OR
       p_batch->>'checkpoint_key' IS DISTINCT FROM p_checkpoint_key OR
       p_batch->>'checkpoint_state' IS DISTINCT FROM p_checkpoint_state OR
       p_batch->'last_key' IS DISTINCT FROM p_last_key OR
       p_batch->'source_position'->>'snapshot_name' IS DISTINCT FROM
           p_snapshot_name OR
       (p_batch->'source_position'->>'snapshot_lsn')::pg_lsn IS DISTINCT FROM
           p_snapshot_lsn OR
       (p_batch->'source_position'->>'received_lsn')::pg_lsn IS DISTINCT FROM
           p_received_lsn OR
       (p_batch->'source_position'->>'commit_lsn')::pg_lsn IS DISTINCT FROM
           p_commit_lsn OR
       (p_batch->'source_position'->>'final_lsn')::pg_lsn IS DISTINCT FROM
           p_final_lsn OR
       (p_batch->'source_position'->>'source_xid')::xid8 IS DISTINCT FROM
           p_source_xid OR
       (p_batch->'source_position'->>'batch_ordinal')::integer IS DISTINCT FROM
           p_batch_ordinal THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing V1 adoption batch evidence differs';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_v1_adoption_batch_assert_v1(
    jsonb, uuid, uuid, text, uuid, smallint, text, text,
    pg_lsn, pg_lsn, pg_lsn, pg_lsn, xid8, integer, text, jsonb, text, text
) FROM PUBLIC;
