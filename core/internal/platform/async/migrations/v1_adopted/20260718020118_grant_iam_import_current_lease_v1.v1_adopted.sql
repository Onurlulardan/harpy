-- +goose Up
GRANT EXECUTE ON FUNCTION platform.platform_async_iam_import_current_lease_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,text
) TO harpy_v2_runtime;
