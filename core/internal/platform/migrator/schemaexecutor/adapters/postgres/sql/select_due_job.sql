SELECT
    job.job_id::text,
    job.tenant_id::text,
    job.owner,
    job.resource_kind,
    job.resource_id::text,
    job.resource_revision,
    job.compiler_version,
    job.plan_contract_version,
    COALESCE(
        pg_catalog.convert_from(job.canonical_plan_bytes, 'UTF8'),
        job.canonical_plan::text
    ),
    job.plan_sha256,
    job.step_count,
    job.next_step_ordinal,
    job.attempt,
    job.generation,
    step.generation
FROM platform.schema_jobs AS job
JOIN platform.schema_job_steps AS step
  ON step.job_id = job.job_id
 AND step.ordinal = job.next_step_ordinal
WHERE job.state = 'queued'
  AND job.next_executor = 'migrator'
  AND job.available_at <= pg_catalog.clock_timestamp()
  AND (job.plan_contract_version = 'forms-schema-plan.v1' OR
       platform.platform_migrator_schema_authority_current_v2(
           job.tenant_id, job.submission_authority_id,
           job.submission_authority_group_id, job.submission_group_generation,
           job.submission_billing_generation, job.submission_forms_generation
       ))
ORDER BY job.available_at, job.created_at, job.job_id
FOR UPDATE SKIP LOCKED
LIMIT 1;
