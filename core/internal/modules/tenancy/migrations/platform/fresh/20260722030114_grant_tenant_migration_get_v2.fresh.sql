-- +goose Up
GRANT EXECUTE ON FUNCTION platform.tenancy_tenant_migration_get_v2(uuid)
TO harpy_runtime;
