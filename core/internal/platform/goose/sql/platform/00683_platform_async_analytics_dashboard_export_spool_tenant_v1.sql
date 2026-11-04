-- +goose Up
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_spool_tenant_v1(
    in_export_id uuid
)
RETURNS uuid
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT NULL::uuid;
$$;

REVOKE ALL ON FUNCTION platform.platform_async_analytics_dashboard_export_spool_tenant_v1(uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platform_async_analytics_dashboard_export_spool_tenant_v1(uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platform_async_analytics_dashboard_export_spool_tenant_v1(uuid);
