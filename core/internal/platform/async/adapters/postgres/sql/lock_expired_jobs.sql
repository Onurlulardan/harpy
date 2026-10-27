WITH recovery_clock AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS observed_at
)
SELECT
    job.id,
    attempt.attempt_id,
    job.lease_owner_id,
    job.lease_generation,
    job.attempts,
    job.max_attempts,
    job.effect_class,
    job.cancel_requested_at IS NOT NULL,
    attempt.effect_started_at,
    attempt.effect_id,
    attempt.effect_request_sha256,
    COALESCE(receipt.receipt_count, 0),
    receipt.receipt_id,
    receipt.effect_status,
    receipt.effect_error_class,
    receipt.effect_code,
    GREATEST(
        recovery_clock.observed_at,
        job.updated_at + interval '1 microsecond'
    ) AS recovered_at
FROM platform.jobs AS job
JOIN platform.job_attempts AS attempt
  ON attempt.job_id = job.id
 AND attempt.attempt = job.attempts
 AND attempt.lease_generation = job.lease_generation
 AND attempt.lease_owner_id = job.lease_owner_id
 AND attempt.finished_at IS NULL
LEFT JOIN LATERAL (
    SELECT
        candidate.receipt_id,
        candidate.effect_status,
        candidate.effect_error_class,
        candidate.effect_code,
        pg_catalog.count(*) OVER () AS receipt_count
    FROM platform.idempotency_receipts AS candidate
    WHERE candidate.record_kind = 'effect'
      AND candidate.scope_kind = job.scope_kind
      AND candidate.scope_id = job.scope_id
      AND candidate.authority_group_id = job.authority_group_id
      AND candidate.authority_group_generation = job.authority_group_generation
      AND candidate.authority_domain = job.authority_domain
      AND candidate.cutover_generation = job.cutover_generation
      AND candidate.writer_origin = 'v2'
      AND candidate.effect_id = attempt.effect_id
      AND candidate.request_sha256 = attempt.effect_request_sha256
      AND candidate.effect_key = job.effect_idempotency_key
    ORDER BY candidate.completed_at DESC, candidate.receipt_id
    LIMIT 1
) AS receipt ON true
CROSS JOIN recovery_clock
WHERE job.scope_kind = $1
  AND job.scope_id = $2
  AND job.authority_group_id = $3
  AND job.authority_group_generation = $4
  AND job.authority_domain = $5
  AND job.cutover_generation = $6
  AND job.state = 'running'
  AND job.kind <> 'iam.system-resource.export-reconcile'
  AND job.kind <> 'analytics.dashboard.export.render'
  AND (
      ($8::text = 'v2' AND NOT EXISTS (
          SELECT 1 FROM platform.cutover_job_bindings AS binding
          WHERE binding.job_id=job.id
      )) OR
      ($8::text = 'cutover_shadow' AND job.kind='documents.adoption.copy'
       AND EXISTS (
          SELECT 1 FROM platform.cutover_job_bindings AS binding
          WHERE binding.job_id=job.id
            AND platform.platform_async_cutover_shadow_operation_active_v1(
                binding.operation_id,binding.cutover_run_id,
                job.scope_kind,job.scope_id,job.authority_group_id,
                job.authority_domain,job.cutover_generation,
                binding.lease_manifest_sha256,recovery_clock.observed_at
            )
       ))
  )
  AND job.lease_until <= recovery_clock.observed_at
ORDER BY job.lease_until, job.id
FOR UPDATE OF job SKIP LOCKED
LIMIT $7;
