-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_iam_import_cleanup_tenant_v1(
    in_after_value uuid
)
RETURNS TABLE (
    tenant_id uuid,
    tenant_schematext text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_iam_import_cleanup_tenant_v1(uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platform_async_iam_import_cleanup_tenant_v1(uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platform_async_iam_import_cleanup_tenant_v1(uuid);
