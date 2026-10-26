INSERT INTO platform.jobs (
    id, tenant_id, scope_kind, scope_id,
    authority_group_id, authority_group_generation,
    authority_domain, cutover_generation, writer_origin,
    kind, payload_revision, payload_ref, payload_sha256, dedupe_key,
    effect_class, effect_idempotency_key, priority, run_at,
    state, attempts, max_attempts, lease_generation, retry_of_job_id
) VALUES (
    $1, $2, $3, $4,
    $5, $6,
    $7, $8, $9,
    $10, $11, $12, $13, $14,
    $15, $16, $17, $18,
    'queued', 0, $19, 0, $20
)
ON CONFLICT DO NOTHING;
