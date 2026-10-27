-- +goose Up
CREATE TRIGGER schema_job_steps_validate_transition
BEFORE UPDATE OR DELETE ON platform.schema_job_steps
FOR EACH ROW
EXECUTE FUNCTION platform.platform_migrator_validate_schema_step_transition();
