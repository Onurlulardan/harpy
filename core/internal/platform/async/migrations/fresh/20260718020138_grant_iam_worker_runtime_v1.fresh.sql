-- +goose Up
GRANT EXECUTE ON FUNCTION platform.platform_async_iam_import_current_lease_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,text
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platform_async_iam_export_current_lease_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,text
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platform_async_iam_export_effect_start_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,text,uuid,uuid,bytea,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platform_async_iam_export_succeed_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,text,uuid,uuid,bytea,bytea,text,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platform_async_iam_export_finish_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,text,text,text,bytea,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platform_async_iam_export_unknown_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,text,uuid,bytea,timestamptz
) TO forgecrud_runtime;
