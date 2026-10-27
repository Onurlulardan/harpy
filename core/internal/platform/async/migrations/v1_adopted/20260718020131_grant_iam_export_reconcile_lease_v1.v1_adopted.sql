-- +goose Up
GRANT EXECUTE ON FUNCTION platform.platform_async_iam_export_reconcile_lease_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,text
) TO harpy_runtime;
