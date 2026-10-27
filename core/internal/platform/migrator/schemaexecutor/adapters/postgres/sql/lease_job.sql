WITH authoritative_time AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS now
)
UPDATE platform.schema_jobs AS job
SET state = 'leased',
    generation = generation + 1,
    attempt = attempt + 1,
    lease_id = $2,
    lease_owner_id = $3,
    leased_at = authoritative_time.now,
    lease_heartbeat_at = authoritative_time.now,
    lease_expires_at = authoritative_time.now + pg_catalog.make_interval(secs => $4),
    updated_at = authoritative_time.now
FROM authoritative_time
WHERE job.job_id = $1
  AND job.state = 'queued'
  AND job.next_executor = 'migrator'
  AND job.available_at <= authoritative_time.now
  AND job.generation = $5
  AND (job.plan_contract_version = 'forms-schema-plan.v1' OR
       platform.platform_migrator_schema_authority_current_v2(
           job.tenant_id, job.submission_authority_id,
           job.submission_authority_group_id, job.submission_group_generation,
           job.submission_billing_generation, job.submission_forms_generation
       ))
RETURNING job.attempt, job.lease_expires_at, job.generation;
