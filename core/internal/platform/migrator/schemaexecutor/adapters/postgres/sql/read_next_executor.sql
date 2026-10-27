SELECT executor_kind
FROM platform.schema_job_steps
WHERE job_id = $1
  AND ordinal = $2
  AND state = 'pending';
