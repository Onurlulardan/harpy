SELECT
    job.job_id::text,
    job.tenant_id::text,
    job.lease_id::text,
    job.lease_owner_id::text,
    job.attempt,
    step.attempt,
    job.next_step_ordinal,
    job.plan_sha256,
    job.lease_expires_at,
    step.step_sha256,
    step.started_at,
    job.generation,
    step.generation
FROM platform.schema_jobs AS job
JOIN platform.schema_job_steps AS step
  ON step.job_id = job.job_id
 AND step.ordinal = job.next_step_ordinal
WHERE job.state IN ('leased', 'running')
  AND job.next_executor = 'migrator'
  AND job.lease_expires_at <= pg_catalog.clock_timestamp()
  AND step.state = 'running'
ORDER BY job.lease_expires_at, job.job_id
FOR UPDATE SKIP LOCKED
LIMIT 1;
