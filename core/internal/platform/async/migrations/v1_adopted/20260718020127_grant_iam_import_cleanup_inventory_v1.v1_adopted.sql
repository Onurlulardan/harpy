-- +goose Up
GRANT EXECUTE ON FUNCTION
    platform.platform_async_iam_import_cleanup_tenant_v1(uuid)
    TO harpy_runtime;

GRANT EXECUTE ON FUNCTION platform.platform_async_iam_import_cleanup_read_v1(
    uuid,uuid,uuid,text
) TO harpy_runtime;
