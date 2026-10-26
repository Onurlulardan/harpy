INSERT INTO platform.billing_subscription_audit (
    id, authority_id, tenant_id, subscription_id, source_operation_id,
    source_kind, kind, actor_id, before_revision, after_revision,
    occurred_at, reason
) VALUES ($3,$1,$2,$4,$5,$6,$7,$8,$9,$10,$11,$12)
RETURNING id;
