-- +goose Up
-- BELIRSIZ: tenant identity is not loaded. not_deleted is false.
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_runtime_resolve_command_v1(
    p_tenant_id uuid
)
RETURNS TABLE (
    schema_name text,
    slug text,
    status text,
    not_deleted boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT ''::text, ''::text, ''::text, false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_runtime_resolve_command_v1(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_runtime_resolve_command_v1(uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_runtime_resolve_command_v1(uuid);
