-- +goose Up
GRANT EXECUTE ON FUNCTION platform.tenancy_tenant_migration_complete_v2(
    uuid, integer, uuid, uuid, uuid, text, text
) TO harpy_tenant_migration_executor;

