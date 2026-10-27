-- +goose Up
GRANT EXECUTE ON FUNCTION platform.platform_async_iam_export_reconcile_effect_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,text,uuid,uuid,bytea,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.platform_async_iam_export_reconcile_sealed_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,text,uuid,uuid,bytea,bytea,text,timestamptz,
    uuid,uuid,uuid,bytea,bytea,text,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.platform_async_iam_export_reconcile_missing_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,text,uuid,uuid,bytea,bytea,text,timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.platform_async_iam_export_reconcile_retry_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,text,uuid,uuid,bytea,bytea,timestamptz,
    timestamptz
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.platform_async_iam_export_reconcile_park_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,text,uuid,uuid,bytea,bytea,timestamptz
) TO harpy_runtime;
