INSERT INTO platform.billing_subscription_operations (
    id, authority_id, tenant_id, subscription_id, source_operation_id,
    source_kind, operation_kind, request_sha256, result_sha256,
    expected_revision, result_revision, result_state, completed_at,
    result_provider_account_id, result_offering_version_id,
    result_price_version_id, result_currency, result_period_start,
    result_period_end, result_trial_ends_at, result_grace_ends_at,
    result_cancel_at_period_end, result_cancel_at, result_updated_at
) VALUES (
    $3,$1,$2,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,
    $14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24
)
RETURNING id;
