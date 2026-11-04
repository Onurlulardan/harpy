-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_installation_assert_v2(
    in_operation_id uuid,
    in_tenant_id uuid,
    in_schema_name name
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

REVOKE ALL ON FUNCTION system_resource_layout_installation_assert_v2(uuid, uuid, name) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION system_resource_layout_installation_assert_v2(uuid, uuid, name) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION system_resource_layout_installation_assert_v2(uuid, uuid, name);
