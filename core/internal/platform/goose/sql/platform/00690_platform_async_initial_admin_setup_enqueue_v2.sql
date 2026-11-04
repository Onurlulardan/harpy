-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_initial_admin_setup_enqueue_v2(
    in_operation_id uuid,
    in_environment_id uuid,
    in_tenant_id uuid,
    in_artifact_id uuid,
    in_artifact_reference text,
    in_schema_name name,
    in_manifest_sha256 text
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_initial_admin_setup_enqueue_v2(uuid, uuid, uuid, uuid, text, name, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platform_async_initial_admin_setup_enqueue_v2(uuid, uuid, uuid, uuid, text, name, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platform_async_initial_admin_setup_enqueue_v2(uuid, uuid, uuid, uuid, text, name, text);
