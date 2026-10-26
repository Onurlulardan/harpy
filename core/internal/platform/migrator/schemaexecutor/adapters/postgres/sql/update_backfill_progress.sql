WITH authoritative_time AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS now
)
UPDATE platform.schema_job_steps AS step
SET generation = generation + 1,
    checkpoint_record_id = $7,
    processed_rows = processed_rows + $8,
    checkpoint_canonical_bytes = $9,
    checkpoint_page_sha256 = $10,
    updated_at = authoritative_time.now
FROM authoritative_time
WHERE step.job_id = $1
  AND step.ordinal = $2
  AND step.state = 'running'
  AND step.attempt = $3
  AND step.plan_sha256 = $4
  AND step.step_sha256 = $5
  AND step.generation = $6
RETURNING step.generation;
