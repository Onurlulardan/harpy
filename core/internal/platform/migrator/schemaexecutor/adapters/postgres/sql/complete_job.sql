WITH authoritative_time AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS now
)
UPDATE platform.schema_jobs AS job
SET state = 'succeeded',
    generation = generation + 1,
    next_step_ordinal = step_count,
    next_executor = 'none',
    lease_id = NULL,
    lease_owner_id = NULL,
    leased_at = NULL,
    lease_heartbeat_at = NULL,
    lease_expires_at = NULL,
    updated_at = authoritative_time.now,
    finished_at = authoritative_time.now
FROM authoritative_time
WHERE job.job_id = $1
  AND job.state = 'running'
  AND job.lease_id = $2
  AND job.lease_owner_id = $3
  AND job.attempt = $4
  AND job.plan_sha256 = $5
  AND job.generation = $6
  AND job.lease_expires_at > authoritative_time.now
RETURNING job.generation;
