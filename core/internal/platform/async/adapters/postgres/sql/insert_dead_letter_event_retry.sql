INSERT INTO platform.jobs (
    id, tenant_id, scope_kind, scope_id,
    authority_group_id, authority_group_generation,
    authority_domain, cutover_generation, writer_origin,
    kind, payload_revision, payload_ref, payload_sha256, dedupe_key,
    effect_class, effect_idempotency_key, priority, run_at,
    state, attempts, max_attempts, lease_generation, retry_of_job_id
)
SELECT
    $2, root.tenant_id, root.scope_kind, root.scope_id,
    root.authority_group_id, root.authority_group_generation,
    root.authority_domain, root.cutover_generation, 'v2',
    $3, 1, $4, root.source_sha256, $5,
    'external', $6, 100, pg_catalog.clock_timestamp(),
    'queued', 0, $7, 0, NULL
FROM platform.dead_letters AS root
WHERE root.id = $1
  AND root.record_kind = 'root'
  AND root.source_kind = 'event'
RETURNING id;
