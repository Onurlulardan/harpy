-- +goose Up
ALTER TABLE platform.tenant_migration_checkpoints
    DISABLE TRIGGER tenant_migration_checkpoints_validate_transition;

CREATE TRIGGER tenant_migration_checkpoints_validate_transition_v2
BEFORE INSERT OR UPDATE OR DELETE ON platform.tenant_migration_checkpoints
FOR EACH ROW
EXECUTE FUNCTION platform.platform_migrator_validate_checkpoint_transition_v2();
