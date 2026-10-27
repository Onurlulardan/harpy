-- +goose Up
GRANT EXECUTE ON FUNCTION
    platform.platform_async_forms_import_source_kek_tenant_v1(uuid)
    TO forgecrud_runtime;
