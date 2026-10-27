-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_authorization_evidence_read_v1(
    p_scope_schema name,
    p_actor_id uuid,
    p_resource_key text,
    p_action_key text
)
RETURNS TABLE (
    actor_document jsonb,
    resource_document jsonb,
    action_document jsonb,
    authorization_clock bigint,
    bindings_document jsonb,
    implications_document jsonb
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_actor_id IS NULL OR p_resource_key IS NULL OR p_action_key IS NULL OR
       pg_catalog.char_length(p_resource_key) NOT BETWEEN 1 AND 100 OR
       pg_catalog.char_length(p_action_key) NOT BETWEEN 1 AND 100 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM authorization evidence scope is invalid';
    END IF;
    RETURN QUERY
    WITH actor AS (
        SELECT id, status, token_version, authorization_version
        FROM iam_users
        WHERE id = p_actor_id
    ), target_resource AS (
        SELECT id, state, is_protected
        FROM iam_resources
        WHERE resource_key = p_resource_key
    ), target_action AS (
        SELECT id, state, is_protected
        FROM iam_actions
        WHERE action_key = p_action_key
    ), target_exact AS (
        SELECT permission.id
        FROM iam_permissions AS permission
        JOIN target_resource AS resource ON resource.id = permission.resource_id
        JOIN target_action AS action ON action.id = permission.action_id
        WHERE permission.match_kind = 'exact'
    ), target_edges AS (
        SELECT implication.*
        FROM iam_permission_implications AS implication
        JOIN target_exact AS target ON target.id = implication.implied_permission_id
    ), relevant_permissions AS (
        SELECT id FROM target_exact
        UNION
        SELECT source_permission_id FROM target_edges
        UNION
        SELECT permission.id
        FROM iam_permissions AS permission
        JOIN target_resource AS resource ON resource.id = permission.resource_id
        WHERE permission.match_kind = 'resource_all_actions'
        UNION
        SELECT id FROM iam_permissions WHERE match_kind = 'tenant_all'
    ), binding_paths AS (
        SELECT
            'user'::text AS path, binding.id AS binding_id,
            binding.state AS binding_state, binding.permission_id,
            NULL::uuid AS assignment_id, ''::text AS assignment_state,
            NULL::uuid AS role_id, ''::text AS role_status,
            ''::text AS role_scope_kind, false AS role_protected,
            NULL::uuid AS membership_id, ''::text AS membership_state,
            NULL::uuid AS organization_id, ''::text AS organization_status,
            NULL::uuid AS policy_role_id, NULL::uuid AS policy_permission_id,
            NULL::uuid AS policy_binding_id, ''::text AS policy_state
        FROM iam_user_permission_bindings AS binding
        JOIN relevant_permissions AS relevant ON relevant.id = binding.permission_id
        WHERE binding.user_id = p_actor_id
        UNION ALL
        SELECT
            'role', binding.id, binding.state, binding.permission_id,
            assignment.id, assignment.state, role.id, role.status,
            role.scope_kind, role.is_protected,
            membership.id, COALESCE(membership.state, ''),
            organization.id, COALESCE(organization.status, ''),
            policy.role_id, policy.tenant_all_permission_id,
            policy.role_permission_binding_id, COALESCE(policy.policy_state, '')
        FROM iam_user_role_assignments AS assignment
        JOIN iam_roles AS role ON role.id = assignment.role_id
        JOIN iam_role_permission_bindings AS binding ON binding.role_id = role.id
        JOIN relevant_permissions AS relevant ON relevant.id = binding.permission_id
        LEFT JOIN iam_organization_memberships AS membership
            ON membership.id = assignment.membership_id
        LEFT JOIN iam_organizations AS organization
            ON organization.id = assignment.organization_id
        LEFT JOIN iam_protected_admin_policies AS policy ON policy.role_id = role.id
        WHERE assignment.user_id = p_actor_id
        UNION ALL
        SELECT
            'organization', binding.id, binding.state, binding.permission_id,
            NULL::uuid, '', NULL::uuid, '', '', false,
            membership.id, membership.state, organization.id, organization.status,
            NULL::uuid, NULL::uuid, NULL::uuid, ''
        FROM iam_organization_memberships AS membership
        JOIN iam_organizations AS organization
          ON organization.id = membership.organization_id
        JOIN iam_organization_permission_bindings AS binding
          ON binding.organization_id = organization.id
        JOIN relevant_permissions AS relevant ON relevant.id = binding.permission_id
        WHERE membership.user_id = p_actor_id
    ), binding_evidence AS (
        SELECT path.*, permission.match_kind, permission.state AS permission_state,
            permission.is_protected AS permission_protected,
            permission.resource_id, permission.action_id,
            COALESCE(resource.state, '') AS resource_state,
            COALESCE(action.state, '') AS action_state
        FROM binding_paths AS path
        JOIN iam_permissions AS permission ON permission.id = path.permission_id
        LEFT JOIN iam_resources AS resource ON resource.id = permission.resource_id
        LEFT JOIN iam_actions AS action ON action.id = permission.action_id
    ), implication_evidence AS (
        SELECT implication.id, implication.implication_kind, implication.state,
            source.id AS source_permission_id,
            source.state AS source_permission_state,
            source.resource_id AS source_resource_id,
            source.action_id AS source_action_id,
            source_resource.state AS source_resource_state,
            source_action.state AS source_action_state,
            implied.id AS implied_permission_id,
            implied.state AS implied_permission_state,
            implied.resource_id AS implied_resource_id,
            implied.action_id AS implied_action_id,
            implied_resource.state AS implied_resource_state,
            implied_action.state AS implied_action_state
        FROM target_edges AS implication
        JOIN iam_permissions AS source ON source.id = implication.source_permission_id
        JOIN iam_resources AS source_resource ON source_resource.id = source.resource_id
        JOIN iam_actions AS source_action ON source_action.id = source.action_id
        JOIN iam_permissions AS implied ON implied.id = implication.implied_permission_id
        JOIN iam_resources AS implied_resource ON implied_resource.id = implied.resource_id
        JOIN iam_actions AS implied_action ON implied_action.id = implied.action_id
    )
    SELECT
        (SELECT pg_catalog.jsonb_build_object(
            'id', id, 'status', status, 'token_version', token_version,
            'authorization_version', authorization_version
        ) FROM actor),
        (SELECT pg_catalog.jsonb_build_object(
            'id', id, 'state', state, 'protected', is_protected
        ) FROM target_resource),
        (SELECT pg_catalog.jsonb_build_object(
            'id', id, 'state', state, 'protected', is_protected
        ) FROM target_action),
        COALESCE((SELECT clock_value FROM iam_authorization_clock WHERE singleton), -1),
        COALESCE((SELECT pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
            'path', path, 'binding_id', binding_id,
            'binding_state', binding_state, 'permission_id', permission_id,
            'permission_match_kind', match_kind, 'permission_state', permission_state,
            'permission_protected', permission_protected,
            'permission_resource', resource_id, 'permission_action', action_id,
            'resource_state', resource_state, 'action_state', action_state,
            'assignment_id', assignment_id, 'assignment_state', assignment_state,
            'role_id', role_id, 'role_status', role_status,
            'role_scope_kind', role_scope_kind, 'role_protected', role_protected,
            'membership_id', membership_id, 'membership_state', membership_state,
            'organization_id', organization_id,
            'organization_status', organization_status,
            'policy_role_id', policy_role_id,
            'policy_permission_id', policy_permission_id,
            'policy_binding_id', policy_binding_id, 'policy_state', policy_state
        ) ORDER BY path, binding_id) FROM binding_evidence), '[]'::jsonb),
        COALESCE((SELECT pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
            'id', id, 'kind', implication_kind, 'state', state,
            'source_permission_id', source_permission_id,
            'source_permission_state', source_permission_state,
            'source_resource_id', source_resource_id,
            'source_action_id', source_action_id,
            'source_resource_state', source_resource_state,
            'source_action_state', source_action_state,
            'implied_permission_id', implied_permission_id,
            'implied_permission_state', implied_permission_state,
            'implied_resource_id', implied_resource_id,
            'implied_action_id', implied_action_id,
            'implied_resource_state', implied_resource_state,
            'implied_action_state', implied_action_state
        ) ORDER BY id) FROM implication_evidence), '[]'::jsonb);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_authorization_evidence_read_v1(name,uuid,text,text)
    FROM PUBLIC;
