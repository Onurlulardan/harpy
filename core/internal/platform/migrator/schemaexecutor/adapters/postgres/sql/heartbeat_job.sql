WITH authoritative_time AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS now
)
UPDATE platform.schema_jobs AS job
SET generation = generation + 1,
    lease_heartbeat_at = authoritative_time.now,
    lease_expires_at = authoritative_time.now + pg_catalog.make_interval(secs => $6),
    updated_at = authoritative_time.now
FROM authoritative_time
WHERE job.job_id = $1
  AND job.state IN ('leased', 'running')
  AND job.lease_id = $2
  AND job.lease_owner_id = $3
  AND job.attempt = $4
  AND job.plan_sha256 = $5
  AND job.lease_expires_at > authoritative_time.now
  AND job.generation = $7
RETURNING job.generation, job.lease_expires_at;
