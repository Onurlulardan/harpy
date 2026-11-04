-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_runtime_resolve_identity_v3(
    in_tenant_id uuid
)
RETURNS TABLE (
    schema_name name,
    slug text,
    status text,
    not_deleted boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        ''::name,
        ''::text,
        ''::text,
        false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_runtime_resolve_identity_v3(uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_runtime_resolve_identity_v3(uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_runtime_resolve_identity_v3(uuid);
