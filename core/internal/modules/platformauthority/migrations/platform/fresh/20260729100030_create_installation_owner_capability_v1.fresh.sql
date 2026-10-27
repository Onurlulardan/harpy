-- +goose Up
-- Migration stream version: 20260729100030.
-- +goose StatementBegin
CREATE FUNCTION platform.platformauthority_installation_owner_bind_v3(
    p_operation_id uuid,
    p_environment_id uuid,
    p_tenant_id uuid,
    p_schema_name name,
    p_deployment_mode text,
    p_instance_owner_binding_id uuid,
    p_protected_admin_binding_id uuid,
    p_registry_binding_id uuid,
    p_manifest_sha256 text
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF NOT platform.tenancy_activation_step_context_assert_v1(
        p_operation_id, 'platformauthority.bind_owner',
        p_environment_id, p_tenant_id, p_schema_name,
        'environment_initialize'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'installation owner capability required';
    END IF;
    RETURN platform.platformauthority_installation_owner_bind_v2(
        p_operation_id, p_environment_id, p_tenant_id, p_schema_name,
        p_deployment_mode, p_instance_owner_binding_id,
        p_protected_admin_binding_id, p_registry_binding_id,
        p_manifest_sha256
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platformauthority_installation_owner_bind_v3(
        uuid,uuid,uuid,name,text,uuid,uuid,uuid,text
    ) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION
    platform.platformauthority_installation_owner_bind_v3(
        uuid,uuid,uuid,name,text,uuid,uuid,uuid,text
    ) TO harpy_runtime;
