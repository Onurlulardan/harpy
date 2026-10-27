-- +goose Up
CREATE TRIGGER schema_job_step_attempts_validate_insert
BEFORE INSERT ON platform.schema_job_step_attempts
FOR EACH ROW
EXECUTE FUNCTION platform.platform_migrator_validate_schema_attempt_insert();
