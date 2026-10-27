SELECT id, discount_version_id, subscription_id, offering_version_id,
       price_version_id, source_operation_id, policy_sha256, redeemed_at
FROM platform.billing_discount_redemptions
WHERE authority_id = $1 AND tenant_id = $2 AND source_operation_id = $3;
