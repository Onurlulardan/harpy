SELECT group_row.generation, tenancy_row.generation
FROM platform.cutover_groups AS group_row
JOIN platform.cutover_group_members AS tenancy_member
  ON tenancy_member.scope_kind = group_row.scope_kind
 AND tenancy_member.scope_id = group_row.scope_id
 AND tenancy_member.group_id = group_row.group_id
 AND tenancy_member.domain = 'tenancy'
 AND tenancy_member.ordinal = 1
JOIN platform.domain_cutovers AS tenancy_row
  ON tenancy_row.scope_kind = tenancy_member.scope_kind
 AND tenancy_row.scope_id = tenancy_member.scope_id
 AND tenancy_row.domain = tenancy_member.domain
WHERE group_row.scope_kind = 'tenant'
  AND group_row.scope_id = $1
  AND group_row.group_id = 'tenancy'
  AND group_row.state = 'v2' AND group_row.writer = 'v2'
  AND tenancy_row.state = 'v2' AND tenancy_row.writer = 'v2'
  AND NOT EXISTS (
    SELECT 1
    FROM platform.cutover_group_members AS extra
    WHERE extra.scope_kind = group_row.scope_kind
      AND extra.scope_id = group_row.scope_id
      AND extra.group_id = group_row.group_id
      AND extra.domain <> 'tenancy'
  );
