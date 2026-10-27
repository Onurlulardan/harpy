INSERT INTO platform.billing_trial_claims (
    id, authority_id, tenant_id, subscription_id,
    offering_version_id, source_operation_id, claimed_at
) VALUES ($3,$1,$2,$4,$5,$6,$7)
RETURNING id;
