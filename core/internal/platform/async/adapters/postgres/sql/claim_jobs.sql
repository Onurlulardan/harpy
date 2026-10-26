WITH claim_clock AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS started_at
), candidate AS MATERIALIZED (
    SELECT
        job.id,
        GREATEST(
            claim_clock.started_at,
            job.updated_at + interval '1 microsecond'
        ) AS started_at
    FROM platform.jobs AS job
    CROSS JOIN claim_clock
    WHERE job.scope_kind = $1
      AND job.scope_id = $2
      AND job.authority_group_id = $3
      AND job.authority_group_generation = $4
      AND job.authority_domain = $5
      AND job.cutover_generation = $6
      AND job.state = 'queued'
      AND job.parked_at IS NULL
      AND job.attempts < job.max_attempts
      AND job.run_at <= claim_clock.started_at
      AND (
          pg_catalog.left(
              job.kind,
              pg_catalog.length('platform_async.event_gap_reconcile.')
          ) <> 'platform_async.event_gap_reconcile.'
          OR platform.platform_async_event_gap_job_ready_v2(job.id)
      )
      AND EXISTS (
          SELECT 1
          FROM pg_catalog.generate_subscripts($9::text[], 1) AS policy(position)
          WHERE ($9::text[])[policy.position] = job.kind
            AND ($10::integer[])[policy.position] = job.payload_revision
            AND ($11::text[])[policy.position] = job.authority_domain
            AND ($12::text[])[policy.position] = job.effect_class
            AND ($13::integer[])[policy.position] = job.max_attempts
      )
      AND (
          ($14::text = 'v2' AND NOT EXISTS (
              SELECT 1 FROM platform.cutover_job_bindings AS binding
              WHERE binding.job_id=job.id
          )) OR
          ($14::text = 'cutover_shadow'
           AND job.kind='documents.adoption.copy'
           AND EXISTS (
              SELECT 1 FROM platform.cutover_job_bindings AS binding
              WHERE binding.job_id=job.id
                AND platform.platform_async_cutover_shadow_operation_active_v1(
                    binding.operation_id,binding.cutover_run_id,
                    job.scope_kind,job.scope_id,job.authority_group_id,
                    job.authority_domain,job.cutover_generation,
                    binding.lease_manifest_sha256,claim_clock.started_at
                )
           ))
      )
    ORDER BY job.priority DESC, job.run_at, job.id
    FOR UPDATE SKIP LOCKED
    LIMIT $8
), claimed AS (
    UPDATE platform.jobs AS job
    SET state = 'running',
        attempts = job.attempts + 1,
        lease_generation = job.lease_generation + 1,
        lease_owner_id = $7,
        lease_heartbeat_at = candidate.started_at,
        lease_until = candidate.started_at + interval '60 seconds',
        updated_at = candidate.started_at
    FROM candidate
    WHERE job.id = candidate.id
    RETURNING
        job.id AS job_id,
        job.tenant_id,
        job.scope_kind,
        job.scope_id,
        job.authority_group_id,
        job.authority_group_generation,
        job.authority_domain,
        job.cutover_generation,
        job.kind,
        job.payload_revision,
        job.payload_ref,
        job.payload_sha256,
        job.effect_class,
        job.effect_idempotency_key,
        job.priority,
        job.run_at,
        job.attempts,
        job.max_attempts,
        job.lease_owner_id,
        job.lease_generation,
        job.lease_heartbeat_at AS started_at,
        job.lease_until
)
SELECT *
FROM claimed
ORDER BY priority DESC, run_at, job_id;
