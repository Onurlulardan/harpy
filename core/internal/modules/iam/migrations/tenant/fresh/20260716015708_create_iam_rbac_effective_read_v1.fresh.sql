-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_effective_permissions_v1(
    p_scope_schema name, p_target_kind text, p_target_id uuid, p_limit integer,
    p_actor_id uuid, p_required_permission_id uuid
)
RETURNS TABLE (
    permission_id uuid, match_kind text,
    resource_id uuid, resource_key text,
    action_id uuid, action_key text,
    source_kind text, source_id uuid
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_limit NOT BETWEEN 1 AND 500 OR
       p_target_kind NOT IN ('USER', 'ROLE', 'ORGANIZATION') OR
       NOT iam_rbac_assignment_target_active_v1(p_target_kind, p_target_id) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM RBAC effective permission input is invalid';
    END IF;
    IF NOT iam_rbac_read_assert_v1(
        p_scope_schema, p_actor_id, p_required_permission_id
    ) THEN
        RETURN;
    END IF;
    RETURN QUERY
    WITH bindings AS (
        SELECT binding.permission_id, 'USER'::text AS source_kind,
               binding.user_id AS source_id
        FROM iam_user_permission_bindings AS binding
        WHERE p_target_kind = 'USER' AND binding.user_id = p_target_id
          AND binding.state = 'active'
        UNION ALL
        SELECT binding.permission_id, 'ROLE'::text, binding.role_id
        FROM iam_role_permission_bindings AS binding
        WHERE binding.state = 'active' AND (
            (p_target_kind = 'ROLE' AND binding.role_id = p_target_id) OR
            (p_target_kind = 'USER' AND EXISTS (
                SELECT 1 FROM iam_user_role_assignments AS membership
                WHERE membership.user_id = p_target_id
                  AND membership.role_id = binding.role_id
                  AND membership.state = 'active'
            ))
        )
        UNION ALL
        SELECT binding.permission_id, 'ORGANIZATION'::text,
               binding.organization_id
        FROM iam_organization_permission_bindings AS binding
        WHERE binding.state = 'active' AND (
            (p_target_kind = 'ORGANIZATION' AND
             binding.organization_id = p_target_id) OR
            (p_target_kind = 'USER' AND EXISTS (
                SELECT 1 FROM iam_organization_memberships AS membership
                WHERE membership.user_id = p_target_id
                  AND membership.organization_id = binding.organization_id
                  AND membership.state = 'active'
            ))
        )
    ), expanded AS (
        SELECT binding.permission_id, binding.source_kind, binding.source_id
        FROM bindings AS binding
        UNION
        SELECT implication.implied_permission_id,
               binding.source_kind, binding.source_id
        FROM bindings AS binding
        JOIN iam_permission_implications AS implication
          ON implication.source_permission_id = binding.permission_id
         AND implication.state = 'active'
    )
    SELECT permission.id, permission.match_kind,
           resource.id, coalesce(resource.resource_key, '*'),
           action.id, coalesce(action.action_key, '*'),
           expanded.source_kind, expanded.source_id
    FROM expanded
    JOIN iam_permissions AS permission
      ON permission.id = expanded.permission_id
     AND permission.state = 'active'
    LEFT JOIN iam_resources AS resource
      ON resource.id = permission.resource_id AND resource.state = 'active'
    LEFT JOIN iam_actions AS action
      ON action.id = permission.action_id AND action.state = 'active'
    WHERE (permission.resource_id IS NULL OR resource.id IS NOT NULL)
      AND (permission.action_id IS NULL OR action.id IS NOT NULL)
    ORDER BY coalesce(resource.resource_key, '*'),
             coalesce(action.action_key, '*'),
             expanded.source_kind, expanded.source_id, permission.id
    LIMIT p_limit;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_rbac_effective_permissions_v1(
    name,text,uuid,integer,uuid,uuid
) FROM PUBLIC;
