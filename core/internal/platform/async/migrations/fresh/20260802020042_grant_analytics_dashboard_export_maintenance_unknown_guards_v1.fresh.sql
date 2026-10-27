-- +goose Up
GRANT EXECUTE ON FUNCTION
    platform.platform_async_validate_analytics_dashboard_export_maintenance_unknown_attempt_v1(
        platform.job_attempts,platform.job_attempts,name,name)
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION
    platform.platform_async_validate_analytics_dashboard_export_maintenance_unknown_job_v1(
        platform.jobs,platform.jobs,name,name)
TO forgecrud_runtime;
