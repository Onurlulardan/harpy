-- +goose Up
GRANT EXECUTE ON FUNCTION platform.platform_async_iam_export_current_lease_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,text
) TO forgecrud_v2_runtime;
