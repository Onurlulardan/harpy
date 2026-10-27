-- +goose Up
GRANT EXECUTE ON FUNCTION platform.platform_async_forms_export_candidate_v1()
    TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platform_async_forms_export_kek_tenant_v1(uuid)
    TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platform_async_forms_export_kek_schema_v1(uuid)
    TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platform_async_forms_export_spool_tenant_v1(uuid)
    TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platform_async_forms_export_checkpoint_v1(
    uuid,uuid,uuid,uuid,bigint,uuid,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platform_async_forms_export_effect_start_v1(
    uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platform_async_forms_export_quota_state_v1(uuid)
    TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platform_async_forms_export_quota_fence_v1(
    uuid,bigint,boolean,text,timestamptz
) TO forgecrud_runtime;
