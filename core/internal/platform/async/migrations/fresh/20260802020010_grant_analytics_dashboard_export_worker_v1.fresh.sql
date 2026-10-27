-- +goose Up
GRANT EXECUTE ON FUNCTION
    platform.platform_async_analytics_dashboard_export_candidate_v1()
TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION
    platform.platform_async_analytics_dashboard_export_spool_tenant_v1(uuid)
TO forgecrud_runtime;
