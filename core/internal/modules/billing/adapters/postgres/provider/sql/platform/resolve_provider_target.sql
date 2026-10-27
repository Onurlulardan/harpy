SELECT mapping.tenant_id,
       mapping.id,
       account.id,
       mapping.subscription_id,
       mapping.invoice_id,
       mapping.revision,
       authority.authority_group_generation,
       authority.billing_generation
FROM platform.billing_provider_accounts AS account
JOIN platform.billing_provider_objects AS mapping
  ON mapping.provider_account_id = account.id
JOIN platform.billing_tenant_authorities AS authority
  ON authority.authority_id = mapping.authority_id
 AND authority.tenant_id = mapping.tenant_id
JOIN platform.cutover_groups AS group_row
  ON group_row.scope_kind = authority.scope_kind
 AND group_row.scope_id = authority.scope_id
 AND group_row.group_id = authority.authority_group_id
JOIN platform.domain_cutovers AS billing_row
  ON billing_row.scope_kind = authority.scope_kind
 AND billing_row.scope_id = authority.scope_id
 AND billing_row.domain = authority.billing_domain
WHERE account.provider_key = $1
  AND account.provider_account_key = $2
  AND account.lifecycle = 'active'
  AND account.supports_webhooks
  AND mapping.object_kind = $3
  AND mapping.provider_object_id = $4
  AND mapping.retired_at IS NULL
  AND authority.scope_kind = 'tenant'
  AND authority.scope_id = mapping.tenant_id
  AND authority.authority_group_id = 'billing_forms'
  AND authority.billing_domain = 'billing'
  AND authority.authority_group_generation = group_row.generation
  AND authority.billing_generation = billing_row.generation
  AND group_row.state = 'v2' AND group_row.writer = 'v2'
  AND billing_row.state = 'v2' AND billing_row.writer = 'v2'
FOR SHARE OF account, mapping, authority, group_row, billing_row;
