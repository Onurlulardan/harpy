-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_installation_protected_admin_v1(
    p_tenant_id uuid,
    p_user_id uuid,
    p_assignment_id uuid,
    p_role_id uuid,
    p_permission_id uuid,
    p_role_permission_id uuid,
    p_role_schema_id uuid,
    p_now timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_role_name text;
    v_role_profile jsonb;
BEGIN
    v_role_name := 'Protected Tenant Admin';
    v_role_profile := '{"display_name":"Protected Tenant Admin"}'::jsonb;
    INSERT INTO iam_roles (
        id, scope_kind, scope_id, organization_id, name, name_key,
        description, status, is_protected, is_default,
        profile_resource_kind, profile_schema_version_id, profile_values,
        profile_sha256, revision, created_at, updated_at, retired_at
    ) VALUES (
        p_role_id, 'tenant', p_tenant_id, NULL, v_role_name,
        pg_catalog.convert_to(v_role_name, 'UTF8'),
        'Non-deletable tenant-wide installation authority',
        'active', true, false, 'role', p_role_schema_id, v_role_profile,
        pg_catalog.sha256(pg_catalog.convert_to(v_role_profile::text, 'UTF8')),
        1, p_now, p_now, NULL
    );
    INSERT INTO iam_permissions (
        id, match_kind, resource_id, action_id, is_protected, state,
        revision, created_at, updated_at, retired_at
    ) VALUES (
        p_permission_id, 'tenant_all', NULL, NULL, true, 'active',
        1, p_now, p_now, NULL
    );
    INSERT INTO iam_role_permission_bindings (
        id, role_id, role_is_protected, permission_id,
        permission_match_kind, state, revision, created_at, updated_at,
        activated_at, revoked_at
    ) VALUES (
        p_role_permission_id, p_role_id, true, p_permission_id,
        'tenant_all', 'active', 1, p_now, p_now, p_now, NULL
    );
    INSERT INTO iam_user_role_assignments (
        id, user_id, role_id, role_scope_kind, role_scope_id,
        organization_id, membership_id, state, revision,
        created_at, updated_at, revoked_at
    ) VALUES (
        p_assignment_id, p_user_id, p_role_id, 'tenant', p_tenant_id,
        NULL, NULL, 'active', 1, p_now, p_now, NULL
    );
    INSERT INTO iam_protected_admin_policies (
        singleton, role_id, role_is_protected, role_status, role_scope_kind,
        tenant_all_permission_id, permission_match_kind,
        permission_is_protected, permission_state,
        role_permission_binding_id, role_binding_state,
        minimum_active_assignments, policy_state, authorization_epoch,
        revision, created_at, updated_at
    ) VALUES (
        true, p_role_id, true, 'active', 'tenant', p_permission_id,
        'tenant_all', true, 'active', p_role_permission_id, 'active',
        1, 'active', 1, 1, p_now, p_now
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_installation_protected_admin_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,uuid,timestamptz
) FROM PUBLIC;
