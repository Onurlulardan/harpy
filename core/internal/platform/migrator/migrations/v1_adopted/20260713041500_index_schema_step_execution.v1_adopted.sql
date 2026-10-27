-- +goose Up
CREATE INDEX schema_job_steps_execution_idx
ON platform.schema_job_steps (job_id, state, ordinal)
WHERE state IN ('pending', 'running', 'retryable', 'lease_expired');
