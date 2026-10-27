-- +goose Up
GRANT EXECUTE ON FUNCTION
    platform.platform_async_event_gap_job_ready_v2(uuid)
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION
    platform.platform_async_reconcile_event_gap_v2(uuid,uuid,uuid,bigint)
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION
    platform.platform_async_purge_dead_letters_v2(integer)
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platform_async_dispose_stale_job_v1(
    uuid,text,uuid,uuid,uuid,uuid,text,uuid,text,bigint,text,bigint,text,text,
    bytea,text,text,text,uuid,uuid,text,text,text
) TO forgecrud_runtime;
