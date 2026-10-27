SELECT
    subscription.tenant_id,
    provider.provider_key,
    authority.authority_group_generation,
    authority.billing_generation
FROM platform.billing_subscriptions AS subscription
JOIN platform.billing_tenant_authorities AS authority
  ON authority.authority_id = subscription.authority_id
 AND authority.tenant_id = subscription.tenant_id
JOIN platform.tenants AS tenant
  ON tenant.id = authority.tenant_id
JOIN platform.cutover_groups AS group_row
  ON group_row.scope_kind = authority.scope_kind
 AND group_row.scope_id = authority.scope_id
 AND group_row.group_id = authority.authority_group_id
 AND group_row.generation = authority.authority_group_generation
JOIN platform.domain_cutovers AS billing_row
  ON billing_row.scope_kind = authority.scope_kind
 AND billing_row.scope_id = authority.scope_id
 AND billing_row.domain = authority.billing_domain
 AND billing_row.generation = authority.billing_generation
JOIN platform.billing_provider_accounts AS provider
  ON provider.id = subscription.provider_account_id
WHERE subscription.id = $1
  AND tenant.deleted_at IS NULL
  AND tenant.status IN ('ACTIVE', 'TRIAL')
  AND authority.scope_kind = 'tenant'
  AND authority.authority_group_id = 'billing'
  AND authority.billing_domain = 'billing'
  AND group_row.state = 'v2' AND group_row.writer = 'v2'
  AND billing_row.state = 'v2' AND billing_row.writer = 'v2'
FOR SHARE OF subscription, authority, tenant, group_row,
    billing_row, provider;
