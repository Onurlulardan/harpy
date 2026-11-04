-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_installation_activation_event_v2(
    in_operation_id uuid,
    in_environment_id uuid,
    in_tenant_id uuid,
    in_actor_id uuid,
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

REVOKE ALL ON FUNCTION platform.platform_async_installation_activation_event_v2(uuid, uuid, uuid, uuid, name, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platform_async_installation_activation_event_v2(uuid, uuid, uuid, uuid, name, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platform_async_installation_activation_event_v2(uuid, uuid, uuid, uuid, name, text);
