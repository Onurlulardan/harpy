-- +goose Up
CREATE TRIGGER schema_jobs_validate_transition
BEFORE UPDATE OR DELETE ON platform.schema_jobs
FOR EACH ROW
EXECUTE FUNCTION platform.platform_migrator_validate_schema_job_transition();
