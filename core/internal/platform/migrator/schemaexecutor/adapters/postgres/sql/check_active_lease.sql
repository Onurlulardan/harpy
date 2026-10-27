SELECT job.tenant_id::text
FROM platform.schema_jobs AS job
JOIN platform.schema_job_steps AS step
  ON step.job_id = job.job_id
 AND step.ordinal = job.next_step_ordinal
WHERE job.job_id = $1
  AND job.state = 'running'
  AND job.lease_id = $2
  AND job.lease_owner_id = $3
  AND job.attempt = $4
  AND job.plan_sha256 = $5
  AND job.lease_expires_at > pg_catalog.clock_timestamp()
  AND step.ordinal = $6
  AND step.state = 'running'
  AND step.attempt = $10
  AND step.step_sha256 = $7
  AND job.generation = $8
  AND step.generation = $9
  AND (job.plan_contract_version = 'forms-schema-plan.v1' OR
       platform.platform_migrator_schema_authority_current_v2(
           job.tenant_id, job.submission_authority_id,
           job.submission_authority_group_id, job.submission_group_generation,
           job.submission_billing_generation, job.submission_forms_generation
       ));
