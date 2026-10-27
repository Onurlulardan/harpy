-- +goose Up
GRANT EXECUTE ON FUNCTION
    platform.platform_async_iam_import_cleanup_tenant_v1(uuid)
    TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION platform.platform_async_iam_import_cleanup_read_v1(
    uuid,uuid,uuid,text
) TO forgecrud_runtime;
