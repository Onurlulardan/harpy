-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_user_has_permission_v1(
    p_user_id uuid,
    p_permission_id uuid
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
RETURN (WITH target AS (
    SELECT permission.id, permission.resource_id, permission.action_id
    FROM iam_permissions AS permission
    JOIN iam_resources AS resource ON resource.id = permission.resource_id
    JOIN iam_actions AS action ON action.id = permission.action_id
    WHERE permission.id OPERATOR(pg_catalog.=) p_permission_id
      AND permission.match_kind OPERATOR(pg_catalog.=) 'exact'
      AND permission.state OPERATOR(pg_catalog.=) 'active'
      AND resource.state OPERATOR(pg_catalog.=) 'active'
      AND action.state OPERATOR(pg_catalog.=) 'active'
), relevant AS (
    SELECT target.id FROM target
    UNION
    SELECT implication.source_permission_id
    FROM iam_permission_implications AS implication
    JOIN target ON target.id = implication.implied_permission_id
    WHERE implication.state OPERATOR(pg_catalog.=) 'active'
    UNION
    SELECT permission.id
    FROM iam_permissions AS permission
    JOIN target ON target.resource_id = permission.resource_id
    WHERE permission.match_kind OPERATOR(pg_catalog.=) 'resource_all_actions'
      AND permission.state OPERATOR(pg_catalog.=) 'active'
    UNION
    SELECT permission.id FROM iam_permissions AS permission
    WHERE permission.match_kind OPERATOR(pg_catalog.=) 'tenant_all'
      AND permission.state OPERATOR(pg_catalog.=) 'active'
), direct_allow AS (
    SELECT binding.id
    FROM iam_user_permission_bindings AS binding
    JOIN relevant ON relevant.id = binding.permission_id
    JOIN iam_permissions AS permission ON permission.id = binding.permission_id
    LEFT JOIN iam_resources AS resource ON resource.id = permission.resource_id
    LEFT JOIN iam_actions AS action ON action.id = permission.action_id
    WHERE binding.user_id OPERATOR(pg_catalog.=) p_user_id
      AND binding.state OPERATOR(pg_catalog.=) 'active'
      AND permission.state OPERATOR(pg_catalog.=) 'active'
      AND (resource.id IS NULL OR resource.state OPERATOR(pg_catalog.=) 'active')
      AND (action.id IS NULL OR action.state OPERATOR(pg_catalog.=) 'active')
), role_allow AS (
    SELECT binding.id
    FROM iam_user_role_assignments AS assignment
    JOIN iam_roles AS role ON role.id = assignment.role_id
    JOIN iam_role_permission_bindings AS binding ON binding.role_id = role.id
    JOIN relevant ON relevant.id = binding.permission_id
    JOIN iam_permissions AS permission ON permission.id = binding.permission_id
    LEFT JOIN iam_organization_memberships AS membership
        ON membership.id = assignment.membership_id
    LEFT JOIN iam_organizations AS organization
        ON organization.id = assignment.organization_id
    LEFT JOIN iam_protected_admin_policies AS policy ON policy.role_id = role.id
    WHERE assignment.user_id OPERATOR(pg_catalog.=) p_user_id
      AND assignment.state OPERATOR(pg_catalog.=) 'active'
      AND role.status OPERATOR(pg_catalog.=) 'active'
      AND binding.state OPERATOR(pg_catalog.=) 'active'
      AND permission.state OPERATOR(pg_catalog.=) 'active'
      AND (
          role.scope_kind OPERATOR(pg_catalog.=) 'tenant' OR
          (role.scope_kind OPERATOR(pg_catalog.=) 'organization' AND
              membership.state OPERATOR(pg_catalog.=) 'active' AND
              organization.status OPERATOR(pg_catalog.=) 'active')
      )
      AND (
          permission.match_kind OPERATOR(pg_catalog.=) 'exact' OR
          (permission.match_kind OPERATOR(pg_catalog.=) 'resource_all_actions' AND
              role.is_protected AND policy.policy_state OPERATOR(pg_catalog.=) 'active') OR
          (permission.match_kind OPERATOR(pg_catalog.=) 'tenant_all' AND
              role.is_protected AND policy.policy_state OPERATOR(pg_catalog.=) 'active' AND
              policy.tenant_all_permission_id = permission.id AND
              policy.role_permission_binding_id = binding.id)
      )
), organization_allow AS (
    SELECT binding.id
    FROM iam_organization_memberships AS membership
    JOIN iam_organizations AS organization ON organization.id = membership.organization_id
    JOIN iam_organization_permission_bindings AS binding
        ON binding.organization_id = organization.id
    JOIN relevant ON relevant.id = binding.permission_id
    JOIN iam_permissions AS permission ON permission.id = binding.permission_id
    WHERE membership.user_id OPERATOR(pg_catalog.=) p_user_id
      AND membership.state OPERATOR(pg_catalog.=) 'active'
      AND organization.status OPERATOR(pg_catalog.=) 'active'
      AND binding.state OPERATOR(pg_catalog.=) 'active'
      AND permission.state OPERATOR(pg_catalog.=) 'active'
)
SELECT EXISTS (
    SELECT 1 FROM iam_users AS principal
    WHERE principal.id OPERATOR(pg_catalog.=) p_user_id
      AND principal.status OPERATOR(pg_catalog.=) 'active'
) AND EXISTS (SELECT 1 FROM target) AND (
    EXISTS (SELECT 1 FROM direct_allow) OR
    EXISTS (SELECT 1 FROM role_allow) OR
    EXISTS (SELECT 1 FROM organization_allow)
));
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_user_has_permission_v1(uuid,uuid) FROM PUBLIC;
