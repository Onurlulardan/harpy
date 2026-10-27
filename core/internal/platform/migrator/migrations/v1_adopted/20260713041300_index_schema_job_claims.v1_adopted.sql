-- +goose Up
CREATE INDEX schema_jobs_claim_idx
ON platform.schema_jobs (next_executor, available_at, created_at, job_id)
WHERE state IN ('queued', 'awaiting_runtime');
