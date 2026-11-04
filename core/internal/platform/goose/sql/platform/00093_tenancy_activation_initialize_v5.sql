-- +goose Up
CREATE FUNCTION platform.tenancy_activation_initialize_v5(
    in_operation_id uuid,
    in_environment_id uuid,
    in_tenant_id uuid,
    in_name text,
    in_slug text,
    in_owner_user_id uuid,
    in_schema_name text,
    in_binding_purpose text,
    in_manifest_sha256 text,
    in_kind text
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.tenancy_activation_initialize_v5(uuid, uuid, uuid, text, text, uuid, text, text, text, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_activation_initialize_v5(uuid, uuid, uuid, text, text, uuid, text, text, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_activation_initialize_v5(uuid, uuid, uuid, text, text, uuid, text, text, text, text);
