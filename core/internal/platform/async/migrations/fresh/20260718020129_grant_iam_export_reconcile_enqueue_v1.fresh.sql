-- +goose Up
GRANT EXECUTE ON FUNCTION platform.platform_async_iam_export_reconcile_enqueue_v1(
    uuid,uuid,uuid,text,timestamptz
) TO harpy_runtime;
