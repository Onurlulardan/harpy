WITH authoritative_time AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS now
)
UPDATE platform.schema_job_steps AS step
SET state = 'running',
    generation = generation + 1,
    attempt = attempt + 1,
    started_at = authoritative_time.now,
    finished_at = NULL,
    error_code = NULL,
    updated_at = authoritative_time.now
FROM authoritative_time
WHERE step.job_id = $1
  AND step.ordinal = $2
  AND step.executor_kind = 'migrator'
  AND step.state IN ('pending', 'retryable', 'lease_expired')
  AND step.generation = $3
RETURNING step.attempt, step.step_sha256, step.started_at, step.generation;
