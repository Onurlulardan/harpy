-- +goose Up
GRANT EXECUTE ON FUNCTION
    platform.platform_async_analytics_dashboard_export_expired_candidate_v1()
TO forgecrud_runtime;
