-- +goose Up
-- Migration stream version: 20260730120001.
-- +goose StatementBegin
CREATE FUNCTION iam_installation_initialize_v3(
    p_operation_id uuid, p_environment_id uuid, p_tenant_id uuid,
    p_user_id uuid, p_organization_id uuid, p_protected_assignment_id uuid,
    p_email text, p_display_name text, p_tenant_name text,
    p_binding_set_sha256 text, p_user_schema_id uuid, p_user_field_id uuid,
    p_organization_schema_id uuid, p_organization_field_id uuid,
    p_role_schema_id uuid, p_role_field_id uuid, p_identifier_id uuid,
    p_access_mode_id uuid, p_membership_id uuid, p_role_id uuid,
    p_permission_id uuid, p_role_permission_id uuid
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_replayed boolean;
BEGIN
    v_replayed := iam_installation_initialize_v2(
        p_operation_id, p_environment_id, p_tenant_id, p_user_id,
        p_organization_id, p_protected_assignment_id, p_email,
        p_display_name, p_tenant_name, p_binding_set_sha256,
        p_user_schema_id, p_user_field_id, p_organization_schema_id,
        p_organization_field_id, p_role_schema_id, p_role_field_id,
        p_identifier_id, p_access_mode_id, p_membership_id, p_role_id,
        p_permission_id, p_role_permission_id
    );
    PERFORM iam_profile_display_name_key_repair_v1('display_name');
    RETURN v_replayed;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_installation_initialize_v3(
    uuid,uuid,uuid,uuid,uuid,uuid,text,text,text,text,uuid,uuid,
    uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_installation_initialize_v3(
    uuid,uuid,uuid,uuid,uuid,uuid,text,text,text,text,uuid,uuid,
    uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid
) TO harpy_runtime;
