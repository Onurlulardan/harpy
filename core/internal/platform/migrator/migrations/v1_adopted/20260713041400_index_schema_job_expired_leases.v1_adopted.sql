-- +goose Up
CREATE INDEX schema_jobs_expired_lease_idx
ON platform.schema_jobs (lease_expires_at, job_id)
WHERE state IN ('leased', 'running');
