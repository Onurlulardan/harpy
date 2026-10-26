INSERT INTO platform.billing_discount_redemptions (
    id, authority_id, tenant_id, subscription_id, discount_version_id,
    offering_version_id, price_version_id, source_operation_id,
    applied_kind, applied_fixed_minor_units, applied_currency,
    applied_currency_registry_version, applied_basis_points,
    applied_priority, applied_duration, applied_repeating_cycles,
    policy_sha256, redeemed_at, relationship_ended_at
)
SELECT $3, authority.authority_id, authority.tenant_id, subscription.id,
       version_row.id, subscription.offering_version_id,
       subscription.price_version_id, $8, version_row.discount_kind,
       version_row.fixed_minor_units, version_row.currency,
       version_row.currency_registry_version, version_row.basis_points,
       version_row.priority, version_row.duration, version_row.repeating_cycles,
       $9, $10, NULL
FROM platform.billing_tenant_authorities AS authority
JOIN platform.billing_subscriptions AS subscription
  ON subscription.authority_id = authority.authority_id
 AND subscription.tenant_id = authority.tenant_id
JOIN platform.billing_discount_versions AS version_row ON version_row.id = $5
JOIN platform.billing_price_versions AS price
  ON price.id = subscription.price_version_id
 AND price.offering_version_id = subscription.offering_version_id
WHERE authority.authority_id = $1 AND authority.tenant_id = $2
  AND subscription.id = $4
  AND subscription.offering_version_id = $6
  AND subscription.price_version_id = $7
  AND pg_catalog.octet_length($9::bytea) = 32
RETURNING id;
