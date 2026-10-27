-- +goose Up
GRANT SELECT ON TABLE platform.cutover_job_bindings TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION platform.platform_async_assert_worker_authority_v1(
    text,uuid,text,bigint,text,bigint,text
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION
platform.platform_async_cutover_shadow_operation_active_v1(
    uuid,uuid,text,uuid,text,text,bigint,text,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platform_async_validate_job_transition_v2(
    platform.jobs,platform.jobs,name,name
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platform_async_validate_job_attempt_v2(
    platform.job_attempts,platform.job_attempts,name,name
) TO forgecrud_runtime;
