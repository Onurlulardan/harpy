-- +goose Up
CREATE TRIGGER tenant_migration_checkpoint_attempts_validate
BEFORE INSERT OR UPDATE OR DELETE ON platform.tenant_migration_checkpoint_attempts
FOR EACH ROW
EXECUTE FUNCTION platform.platform_migrator_validate_tenant_checkpoint_attempt();
