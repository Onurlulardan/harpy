WITH authoritative_time AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS now
)
UPDATE platform.schema_jobs AS job
SET state = 'lease_expired',
    generation = generation + 1,
    last_lease_expired_at = authoritative_time.now,
    updated_at = authoritative_time.now,
    failure_code = 'lease_expired'
FROM authoritative_time
WHERE job.job_id = $1
  AND job.state IN ('leased', 'running')
  AND job.lease_id = $2
  AND job.lease_owner_id = $3
  AND job.attempt = $4
  AND job.plan_sha256 = $5
  AND job.generation = $6
  AND job.lease_expires_at <= authoritative_time.now
RETURNING job.generation;
