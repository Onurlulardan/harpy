-- +goose Up
CREATE TRIGGER tenants_guard_identity_and_delete
BEFORE UPDATE OR DELETE ON platform.tenants
FOR EACH ROW EXECUTE FUNCTION platform.tenancy_guard_tenant_registry();
