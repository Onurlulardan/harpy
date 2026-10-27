-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_analytics_role_permission_v1(
    p_scope_schema name,
    p_tenant_id uuid,
    p_role_id uuid,
    p_resource_key text,
    p_action_key text
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_group_generation bigint;
    v_iam_generation bigint;
    v_layout_generation bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_id IS NULL OR p_role_id IS NULL OR
       p_resource_key IS NULL OR p_action_key IS NULL OR
       pg_catalog.char_length(p_resource_key) NOT BETWEEN 1 AND 100 OR
       pg_catalog.char_length(p_action_key) NOT BETWEEN 1 AND 100 OR
       NOT EXISTS (
           SELECT 1 FROM tenancy_schema_identity AS identity_row
           WHERE identity_row.singleton AND identity_row.tenant_id = p_tenant_id
             AND identity_row.schema_name = p_scope_schema::text
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM Analytics role permission scope is invalid';
    END IF;

    SELECT authority.generation INTO v_group_generation
    FROM platform.cutover_groups AS authority
    WHERE authority.scope_kind = 'tenant' AND authority.scope_id = p_tenant_id
      AND authority.group_id = 'iam_system_resources'
      AND authority.state = 'v2' AND authority.writer = 'v2';
    SELECT authority.generation INTO v_iam_generation
    FROM platform.domain_cutovers AS authority
    WHERE authority.scope_kind = 'tenant' AND authority.scope_id = p_tenant_id
      AND authority.domain = 'iam'
      AND authority.state = 'v2' AND authority.writer = 'v2';
    SELECT authority.generation INTO v_layout_generation
    FROM platform.domain_cutovers AS authority
    WHERE authority.scope_kind = 'tenant' AND authority.scope_id = p_tenant_id
      AND authority.domain = 'system_resource_layout'
      AND authority.state = 'v2' AND authority.writer = 'v2';
    IF v_group_generation IS NULL OR v_group_generation < 1 OR
       v_iam_generation IS DISTINCT FROM v_group_generation OR
       v_layout_generation IS DISTINCT FROM v_group_generation OR
       NOT EXISTS (
           SELECT 1 FROM platform.cutover_group_members AS member
           WHERE member.scope_kind = 'tenant' AND member.scope_id = p_tenant_id
             AND member.group_id = 'iam_system_resources'
             AND member.ordinal = 1 AND member.domain = 'iam'
       ) OR NOT EXISTS (
           SELECT 1 FROM platform.cutover_group_members AS member
           WHERE member.scope_kind = 'tenant' AND member.scope_id = p_tenant_id
             AND member.group_id = 'iam_system_resources'
             AND member.ordinal = 2 AND member.domain = 'system_resource_layout'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5701',
            MESSAGE = 'IAM Analytics role permission authority is unavailable';
    END IF;

    RETURN EXISTS (
        WITH target AS (
            SELECT permission.id, permission.resource_id, permission.action_id
            FROM iam_permissions AS permission
            JOIN iam_resources AS resource ON resource.id = permission.resource_id
            JOIN iam_actions AS action ON action.id = permission.action_id
            WHERE permission.match_kind = 'exact' AND permission.state = 'active'
              AND resource.resource_key = p_resource_key AND resource.state = 'active'
              AND action.action_key = p_action_key AND action.state = 'active'
        ), eligible AS (
            SELECT target.id FROM target
            UNION ALL
            SELECT permission.id
            FROM iam_permissions AS permission
            JOIN target ON target.resource_id = permission.resource_id
            WHERE permission.match_kind = 'resource_all_actions'
              AND permission.state = 'active'
            UNION ALL
            SELECT permission.id FROM iam_permissions AS permission
            WHERE permission.match_kind = 'tenant_all' AND permission.state = 'active'
        )
        SELECT 1
        FROM iam_roles AS role
        JOIN iam_role_permission_bindings AS binding ON binding.role_id = role.id
        JOIN iam_permissions AS permission ON permission.id = binding.permission_id
        JOIN eligible ON eligible.id = binding.permission_id
        LEFT JOIN iam_protected_admin_policies AS policy ON policy.role_id = role.id
        WHERE role.id = p_role_id AND role.status = 'active'
          AND binding.state = 'active'
          AND (
              permission.match_kind = 'exact' OR
              (permission.match_kind = 'resource_all_actions' AND role.is_protected
               AND policy.policy_state = 'active') OR
              (permission.match_kind = 'tenant_all' AND role.is_protected
               AND policy.policy_state = 'active'
               AND policy.tenant_all_permission_id = permission.id
               AND policy.role_permission_binding_id = binding.id)
          )
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_analytics_role_permission_v1(name,uuid,uuid,text,text)
    FROM PUBLIC;
