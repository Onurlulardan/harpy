INSERT INTO platform.billing_subscription_boundary_jobs (
    id, authority_id, tenant_id, subscription_id, effect_id,
    boundary_kind, run_at, expected_revision, payload_revision,
    payload_reference, payload_sha256, created_at
) VALUES ($3,$1,$2,$4,$5,$6,$7,$8,$9,$10,$11,$12)
RETURNING id;
