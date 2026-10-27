-- +goose Up
GRANT EXECUTE ON FUNCTION platform.platform_async_forms_import_submit_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,text,text,timestamptz
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION platform.platform_async_forms_import_read_v1(
    uuid,uuid,uuid
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION platform.platform_async_forms_import_control_v1(
    uuid,uuid,uuid,boolean,timestamptz
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION platform.platform_async_forms_import_current_lease_v1(
    uuid,uuid,uuid,uuid,uuid,bigint
) TO harpy_runtime;

GRANT EXECUTE ON FUNCTION platform.platform_async_forms_import_complete_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,timestamptz
) TO harpy_runtime;
