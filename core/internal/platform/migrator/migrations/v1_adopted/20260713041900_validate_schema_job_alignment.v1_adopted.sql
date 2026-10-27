-- +goose Up
CREATE CONSTRAINT TRIGGER schema_jobs_validate_alignment
AFTER UPDATE ON platform.schema_jobs
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW
EXECUTE FUNCTION platform.platform_migrator_validate_schema_job_alignment();
