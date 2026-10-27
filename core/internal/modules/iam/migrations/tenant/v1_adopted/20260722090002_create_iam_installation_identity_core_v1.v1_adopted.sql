-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_installation_identity_core_v1(
    p_operation_id uuid,
    p_environment_id uuid,
    p_tenant_id uuid,
    p_user_id uuid,
    p_organization_id uuid,
    p_email text,
    p_display_name text,
    p_tenant_name text,
    p_user_schema_id uuid,
    p_user_field_id uuid,
    p_organization_schema_id uuid,
    p_organization_field_id uuid,
    p_role_schema_id uuid,
    p_role_field_id uuid,
    p_identifier_id uuid,
    p_access_mode_id uuid,
    p_membership_id uuid,
    p_now timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_user_profile jsonb;
    v_organization_profile jsonb;
BEGIN
    v_user_profile := pg_catalog.jsonb_build_object('display_name', p_display_name);
    v_organization_profile := pg_catalog.jsonb_build_object('display_name', p_tenant_name);
    PERFORM iam_installation_profile_v1(
        p_user_schema_id, p_user_field_id, 'user', p_now
    );
    PERFORM iam_installation_profile_v1(
        p_organization_schema_id, p_organization_field_id, 'organization', p_now
    );
    PERFORM iam_installation_profile_v1(
        p_role_schema_id, p_role_field_id, 'role', p_now
    );
    INSERT INTO iam_users (
        id, status, profile_resource_kind, profile_schema_version_id,
        profile_values, profile_sha256, token_version, authorization_version,
        revision, created_at, updated_at, tombstoned_at, identity_kind
    ) VALUES (
        p_user_id, 'active', 'user', p_user_schema_id, v_user_profile,
        pg_catalog.sha256(pg_catalog.convert_to(v_user_profile::text, 'UTF8')),
        1, 1, 1, p_now, p_now, NULL, 'human'
    );
    INSERT INTO iam_user_access_modes (
        id, user_id, mode, state, revision, created_at, updated_at,
        activated_at, suspended_at, retired_at
    ) VALUES (
        p_access_mode_id, p_user_id, 'operator_console', 'active', 1,
        p_now, p_now, p_now, NULL, NULL
    );
    INSERT INTO iam_user_identifiers (
        id, user_id, kind, display_value, normalized_value,
        normalized_sha256, state, revision, verified_at, activated_at,
        tombstoned_at, cooldown_until, retired_at, created_at, updated_at
    ) VALUES (
        p_identifier_id, p_user_id, 'email_v1', p_email, p_email,
        pg_catalog.sha256(pg_catalog.convert_to(p_email, 'UTF8')),
        'active', 1, p_now, p_now, NULL, NULL, NULL, p_now, p_now
    );
    INSERT INTO iam_organizations (
        id, name, slug, status, parent_id, profile_resource_kind,
        profile_schema_version_id, profile_values, profile_sha256,
        revision, created_at, updated_at, tombstoned_at
    ) VALUES (
        p_organization_id, p_tenant_name, 'root', 'active', NULL,
        'organization', p_organization_schema_id, v_organization_profile,
        pg_catalog.sha256(pg_catalog.convert_to(v_organization_profile::text, 'UTF8')),
        1, p_now, p_now, NULL
    );
    INSERT INTO iam_organization_memberships (
        id, organization_id, user_id, membership_kind, state, revision,
        created_at, updated_at, tombstoned_at
    ) VALUES (
        p_membership_id, p_organization_id, p_user_id, 'owner', 'active',
        1, p_now, p_now, NULL
    );
    INSERT INTO iam_authorization_clock (
        singleton, clock_value, revision, created_at, updated_at
    ) VALUES (true, 1, 1, p_now, p_now);
    PERFORM iam_installation_schedules_v1(p_operation_id, p_now);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_installation_identity_core_v1(
    uuid,uuid,uuid,uuid,uuid,text,text,text,uuid,uuid,uuid,uuid,
    uuid,uuid,uuid,uuid,uuid,timestamptz
) FROM PUBLIC;
