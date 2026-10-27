-- +goose Up
CREATE TRIGGER tenant_migration_checkpoints_validate_transition
BEFORE UPDATE OR DELETE ON platform.tenant_migration_checkpoints
FOR EACH ROW EXECUTE FUNCTION platform.platform_migrator_validate_checkpoint_transition();
