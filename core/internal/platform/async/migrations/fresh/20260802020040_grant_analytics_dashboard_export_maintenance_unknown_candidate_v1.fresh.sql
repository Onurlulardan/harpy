-- +goose Up
GRANT EXECUTE ON FUNCTION
    platform.platform_async_analytics_dashboard_export_maintenance_unknown_candidate_v1()
TO harpy_runtime;
