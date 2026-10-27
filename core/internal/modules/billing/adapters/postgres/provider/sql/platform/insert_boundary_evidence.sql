INSERT INTO platform.billing_subscription_boundary_evidence (
    id, authority_id, tenant_id, subscription_id,
    boundary_job_id, provider_event_id, boundary_kind,
    due_at, expected_revision, outcome,
    next_period_start, next_period_end, grace_ends_at, recorded_at
) SELECT
    $1, $2, $3, $4,
    $5, $6, $7,
    $8, $9, $10,
    $11, $12, $13, $14
FROM platform.billing_subscription_boundary_jobs AS intent
WHERE intent.id = $5
  AND intent.authority_id = $2
  AND intent.tenant_id = $3
  AND intent.subscription_id = $4
  AND intent.boundary_kind = $7
  AND intent.run_at = $8
  AND intent.expected_revision = $9
RETURNING id;
