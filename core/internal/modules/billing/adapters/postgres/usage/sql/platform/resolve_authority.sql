SELECT authority.authority_id
FROM platform.billing_tenant_authorities AS authority
JOIN platform.cutover_groups AS group_row
  ON group_row.scope_kind = authority.scope_kind
 AND group_row.scope_id = authority.scope_id
 AND group_row.group_id = authority.authority_group_id
JOIN platform.domain_cutovers AS billing_row
  ON billing_row.scope_kind = authority.scope_kind
 AND billing_row.scope_id = authority.scope_id
 AND billing_row.domain = authority.billing_domain
WHERE authority.tenant_id = $1
  AND authority.authority_group_generation = group_row.generation
  AND authority.billing_generation = billing_row.generation
  AND group_row.state = 'v2' AND group_row.writer = 'v2'
  AND billing_row.state = 'v2' AND billing_row.writer = 'v2'
FOR SHARE OF authority, group_row, billing_row;
