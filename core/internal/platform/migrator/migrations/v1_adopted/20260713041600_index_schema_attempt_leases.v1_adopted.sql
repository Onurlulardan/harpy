-- +goose Up
CREATE INDEX schema_job_step_attempts_lease_idx
ON platform.schema_job_step_attempts (lease_id, job_id, step_ordinal, attempt);
