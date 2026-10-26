SELECT
    job.id,
    job.tenant_id,
    job.scope_kind,
    job.scope_id,
    job.authority_group_id,
    job.authority_group_generation,
    job.authority_domain,
    job.cutover_generation,
    job.writer_origin,
    job.kind,
    job.payload_revision,
    job.payload_ref,
    job.payload_sha256,
    job.dedupe_key,
    job.effect_class,
    job.effect_idempotency_key,
    job.priority,
    job.run_at,
    job.max_attempts,
    job.state,
    job.parked_at,
    job.parked_code,
    job.effect_started_at,
    pg_catalog.transaction_timestamp()
FROM platform.jobs AS job
WHERE job.id = $1::uuid
  AND job.scope_kind = $2::text
  AND job.scope_id = $3::uuid
  AND job.authority_group_id = $4::text
  AND job.authority_domain = $5::text;
