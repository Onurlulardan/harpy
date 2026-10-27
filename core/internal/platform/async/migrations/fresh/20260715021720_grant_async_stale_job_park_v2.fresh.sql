-- +goose Up
GRANT EXECUTE ON FUNCTION platform.platform_async_park_stale_work_v2(
    uuid,text,uuid,bigint,uuid
) TO harpy_runtime;
