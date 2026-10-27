SELECT group_row.generation,
       iam_row.generation,
       layout_row.generation
FROM platform.cutover_groups AS group_row
JOIN platform.cutover_group_members AS iam_member
  ON iam_member.scope_kind = group_row.scope_kind
 AND iam_member.scope_id = group_row.scope_id
 AND iam_member.group_id = group_row.group_id
 AND iam_member.domain = 'iam'
 AND iam_member.ordinal = 1
JOIN platform.cutover_group_members AS layout_member
  ON layout_member.scope_kind = group_row.scope_kind
 AND layout_member.scope_id = group_row.scope_id
 AND layout_member.group_id = group_row.group_id
 AND layout_member.domain = 'system_resource_layout'
 AND layout_member.ordinal = 2
JOIN platform.domain_cutovers AS iam_row
  ON iam_row.scope_kind = iam_member.scope_kind
 AND iam_row.scope_id = iam_member.scope_id
 AND iam_row.domain = iam_member.domain
JOIN platform.domain_cutovers AS layout_row
  ON layout_row.scope_kind = layout_member.scope_kind
 AND layout_row.scope_id = layout_member.scope_id
 AND layout_row.domain = layout_member.domain
WHERE group_row.scope_kind = 'tenant'
  AND group_row.scope_id = $1
  AND group_row.group_id = 'iam_system_resources'
  AND group_row.state = 'v2' AND group_row.writer = 'v2'
  AND iam_row.state = 'v2' AND iam_row.writer = 'v2'
  AND layout_row.state = 'v2' AND layout_row.writer = 'v2'
  AND NOT EXISTS (
    SELECT 1
    FROM platform.cutover_group_members AS extra
    WHERE extra.scope_kind = group_row.scope_kind
      AND extra.scope_id = group_row.scope_id
      AND extra.group_id = group_row.group_id
      AND extra.domain NOT IN ('iam', 'system_resource_layout')
  );
