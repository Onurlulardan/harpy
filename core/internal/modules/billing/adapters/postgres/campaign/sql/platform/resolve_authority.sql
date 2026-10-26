SELECT authority.authority_id
FROM platform.billing_global_authorities AS authority
JOIN platform.cutover_groups AS group_row
  ON group_row.scope_kind = authority.scope_kind
 AND group_row.scope_id = authority.scope_id
 AND group_row.group_id = authority.authority_group_id
JOIN platform.domain_cutovers AS domain_row
  ON domain_row.scope_kind = authority.scope_kind
 AND domain_row.scope_id = authority.scope_id
 AND domain_row.domain = authority.authority_domain
WHERE authority.scope_id = (
    SELECT environment_id FROM platform.environment_identity
)
  AND authority.authority_group_generation = group_row.generation
  AND authority.authority_domain_generation = domain_row.generation
  AND group_row.writer = 'v2' AND group_row.state = 'v2'
  AND domain_row.writer = 'v2' AND domain_row.state = 'v2';
