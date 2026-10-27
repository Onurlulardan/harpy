UPDATE platform.billing_subscriptions AS subscription
SET offering_version_id = $5,
    price_version_id = $6,
    state = $7,
    period_start = $8,
    period_end = $9,
    trial_ends_at = $10,
    grace_ends_at = $11,
    cancel_at_period_end = $12,
    cancel_at = $13,
    updated_at = $14,
    revision = $15
WHERE subscription.authority_id = $1 AND subscription.tenant_id = $2
  AND subscription.id = $3 AND subscription.revision = $4
RETURNING subscription.id;
