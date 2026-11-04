-- +goose Up
CREATE FUNCTION platform.platformauthority_installation_owner_bind_v3(
    in_operation_id uuid,
    in_environment_id uuid,
    in_tenant_id uuid,
    in_schema_name name,
    in_deployment_mode text,
    in_instance_owner_binding_id uuid,
    in_protected_admin_binding_id uuid,
    in_registry_binding_id uuid,
    in_manifest_sha256 text
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.platformauthority_installation_owner_bind_v3(uuid, uuid, uuid, name, text, uuid, uuid, uuid, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformauthority_installation_owner_bind_v3(uuid, uuid, uuid, name, text, uuid, uuid, uuid, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformauthority_installation_owner_bind_v3(uuid, uuid, uuid, name, text, uuid, uuid, uuid, text);
