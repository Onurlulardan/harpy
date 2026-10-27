UPDATE platform.schema_job_steps AS step
SET state = $7,
    generation = generation + 1,
    next_executor = $9,
    finished_at = $10,
    error_code = $8,
    updated_at = $10
WHERE step.job_id = $1
  AND step.ordinal = $2
  AND step.state = 'running'
  AND step.attempt = $3
  AND step.plan_sha256 = $4
  AND step.step_sha256 = $5
  AND step.generation = $6
RETURNING step.generation;
