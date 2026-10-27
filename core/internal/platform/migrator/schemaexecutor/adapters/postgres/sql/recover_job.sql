WITH authoritative_time AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS now
)
UPDATE platform.schema_jobs AS job
SET state = 'queued',
    generation = generation + 1,
    available_at = authoritative_time.now,
    lease_id = NULL,
    lease_owner_id = NULL,
    leased_at = NULL,
    lease_heartbeat_at = NULL,
    lease_expires_at = NULL,
    updated_at = authoritative_time.now,
    failure_code = NULL
FROM authoritative_time
WHERE job.job_id = $1
  AND job.state = 'lease_expired'
  AND job.lease_id = $2
  AND job.lease_owner_id = $3
  AND job.attempt = $4
  AND job.plan_sha256 = $5
  AND job.generation = $6
RETURNING job.generation;
