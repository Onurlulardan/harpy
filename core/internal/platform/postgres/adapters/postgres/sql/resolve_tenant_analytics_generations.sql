SELECT group_row.generation, analytics_row.generation
FROM platform.cutover_groups AS group_row
JOIN platform.cutover_group_members AS analytics_member
  ON analytics_member.scope_kind = group_row.scope_kind
 AND analytics_member.scope_id = group_row.scope_id
 AND analytics_member.group_id = group_row.group_id
 AND analytics_member.domain = 'analytics'
 AND analytics_member.ordinal = 1
JOIN platform.domain_cutovers AS analytics_row
  ON analytics_row.scope_kind = analytics_member.scope_kind
 AND analytics_row.scope_id = analytics_member.scope_id
 AND analytics_row.domain = analytics_member.domain
WHERE group_row.scope_kind = 'tenant'
  AND group_row.scope_id = $1
  AND group_row.group_id = 'analytics'
  AND group_row.state = 'v2' AND group_row.writer = 'v2'
  AND analytics_row.state = 'v2' AND analytics_row.writer = 'v2'
  AND NOT EXISTS (
    SELECT 1
    FROM platform.cutover_group_members AS extra
    WHERE extra.scope_kind = group_row.scope_kind
      AND extra.scope_id = group_row.scope_id
      AND extra.group_id = group_row.group_id
      AND extra.domain <> 'analytics'
  );
