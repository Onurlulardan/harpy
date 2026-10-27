-- +goose Up
GRANT EXECUTE ON FUNCTION platform.platform_async_iam_export_reconcile_expired_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,text,uuid,uuid,bytea,timestamptz,
    timestamptz,timestamptz,text
) TO harpy_runtime;
