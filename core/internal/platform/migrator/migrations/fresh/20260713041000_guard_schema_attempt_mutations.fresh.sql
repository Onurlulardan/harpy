-- +goose Up
CREATE TRIGGER schema_job_step_attempts_reject_mutation
BEFORE UPDATE OR DELETE ON platform.schema_job_step_attempts
FOR EACH ROW
EXECUTE FUNCTION platform.platform_migrator_reject_mutation();
