-- +goose Up
CREATE FUNCTION platform.platformops_installation_recovery_initialize_v2(
    in_operation_id uuid,
    in_environment_id uuid,
    in_tenant_id uuid,
    in_schema_name name,
    in_deployment_mode text,
    in_signing_generation bigint,
    in_trust_bundle_sha256 text,
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

REVOKE ALL ON FUNCTION platform.platformops_installation_recovery_initialize_v2(uuid, uuid, uuid, name, text, bigint, text, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_installation_recovery_initialize_v2(uuid, uuid, uuid, name, text, bigint, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_installation_recovery_initialize_v2(uuid, uuid, uuid, name, text, bigint, text, text);
