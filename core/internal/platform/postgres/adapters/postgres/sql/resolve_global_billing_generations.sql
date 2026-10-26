SELECT group_row.generation, domain_row.generation
FROM platform.environment_identity AS identity
JOIN platform.cutover_groups AS group_row
  ON group_row.scope_kind = 'global'
 AND group_row.scope_id = identity.environment_id
 AND group_row.group_id = 'billing'
JOIN platform.cutover_group_members AS member_row
  ON member_row.scope_kind = group_row.scope_kind
 AND member_row.scope_id = group_row.scope_id
 AND member_row.group_id = group_row.group_id
 AND member_row.domain = 'billing'
JOIN platform.domain_cutovers AS domain_row
  ON domain_row.scope_kind = member_row.scope_kind
 AND domain_row.scope_id = member_row.scope_id
 AND domain_row.domain = member_row.domain
WHERE group_row.state = 'v2' AND group_row.writer = 'v2'
  AND domain_row.state = 'v2' AND domain_row.writer = 'v2'
  AND NOT EXISTS (
    SELECT 1 FROM platform.cutover_group_members AS extra
    WHERE extra.scope_kind = group_row.scope_kind
      AND extra.scope_id = group_row.scope_id
      AND extra.group_id = group_row.group_id
      AND extra.domain <> 'billing'
  );
