-- +goose Up
CREATE UNIQUE INDEX schema_jobs_active_resource_idx
ON platform.schema_jobs (tenant_id, owner, resource_kind, resource_id)
WHERE state NOT IN ('succeeded', 'failed', 'cancelled');
