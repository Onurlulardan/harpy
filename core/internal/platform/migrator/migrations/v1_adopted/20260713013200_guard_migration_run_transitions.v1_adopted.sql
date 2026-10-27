-- +goose Up
CREATE TRIGGER migration_runs_validate_transition
BEFORE UPDATE OR DELETE ON platform.migration_runs
FOR EACH ROW EXECUTE FUNCTION platform.platform_migrator_validate_run_transition();
