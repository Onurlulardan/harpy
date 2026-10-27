INSERT INTO platform.billing_subscription_changes (
    id, authority_id, tenant_id, subscription_id,
    before_offering_version_id, before_price_version_id,
    after_offering_version_id, after_price_version_id,
    source, state, effective_at, source_operation_id,
    decision_id, created_at, resolved_at, revision
) VALUES (
    $3,$1,$2,$4,$5,$6,$7,$8,'manual','scheduled',$9,$10,$11,$12,NULL,$13
)
RETURNING id;
