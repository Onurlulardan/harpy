SELECT subscription.id, subscription.tenant_id,
       subscription.provider_account_id, subscription.offering_version_id,
       subscription.price_version_id, price.currency::text,
       subscription.state, subscription.period_start, subscription.period_end,
       subscription.trial_ends_at, subscription.grace_ends_at,
       subscription.cancel_at_period_end, subscription.cancel_at,
       subscription.revision, subscription.updated_at
FROM platform.billing_tenant_authorities AS authority
JOIN platform.billing_subscriptions AS subscription
  ON subscription.authority_id = authority.authority_id
 AND subscription.tenant_id = authority.tenant_id
JOIN platform.billing_price_versions AS price
  ON price.id = subscription.price_version_id
 AND price.offering_version_id = subscription.offering_version_id
WHERE authority.authority_id = $1 AND authority.tenant_id = $2
  AND subscription.id = $3
FOR UPDATE OF subscription;
