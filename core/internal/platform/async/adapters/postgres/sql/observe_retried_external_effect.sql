SELECT GREATEST(
    pg_catalog.clock_timestamp(),
    job.updated_at,
    attempt.started_at
) AS effect_started_at
FROM platform.jobs AS job
JOIN platform.job_attempts AS attempt
  ON attempt.job_id = job.id
 AND attempt.attempt = job.attempts
WHERE job.scope_kind = $1
  AND job.scope_id = $2
  AND job.authority_group_id = $3
  AND job.authority_group_generation = $4
  AND job.authority_domain = $5
  AND job.cutover_generation = $6
  AND job.id = $7
  AND job.state = 'running'
  AND job.effect_class = 'external'
  AND job.effect_started_at IS NOT NULL
  AND job.cancel_requested_at IS NULL
  AND job.lease_owner_id = $8
  AND job.lease_generation = $9
  AND job.lease_until > pg_catalog.clock_timestamp()
  AND attempt.lease_owner_id = $8
  AND attempt.lease_generation = $9
  AND attempt.finished_at IS NULL
  AND attempt.effect_started_at IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM platform.idempotency_receipts AS receipt
      WHERE receipt.record_kind = 'effect'
        AND receipt.scope_kind = job.scope_kind
        AND receipt.scope_id = job.scope_id
        AND receipt.authority_group_id = job.authority_group_id
        AND receipt.authority_group_generation = job.authority_group_generation
        AND receipt.authority_domain = job.authority_domain
        AND receipt.cutover_generation = job.cutover_generation
        AND receipt.effect_key = job.effect_idempotency_key
        AND receipt.effect_status = 'succeeded'
  );
