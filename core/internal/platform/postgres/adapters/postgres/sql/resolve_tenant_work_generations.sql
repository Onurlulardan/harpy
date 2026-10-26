SELECT group_row.generation, work_row.generation
FROM platform.cutover_groups AS group_row
JOIN platform.cutover_group_members AS work_member
  ON work_member.scope_kind = group_row.scope_kind
 AND work_member.scope_id = group_row.scope_id
 AND work_member.group_id = group_row.group_id
 AND work_member.domain = 'work'
 AND work_member.ordinal = 1
JOIN platform.domain_cutovers AS work_row
  ON work_row.scope_kind = work_member.scope_kind
 AND work_row.scope_id = work_member.scope_id
 AND work_row.domain = work_member.domain
WHERE group_row.scope_kind = 'tenant'
  AND group_row.scope_id = $1
  AND group_row.group_id = 'work'
  AND group_row.state = 'v2' AND group_row.writer = 'v2'
  AND work_row.state = 'v2' AND work_row.writer = 'v2'
  AND NOT EXISTS (
    SELECT 1
    FROM platform.cutover_group_members AS extra
    WHERE extra.scope_kind = group_row.scope_kind
      AND extra.scope_id = group_row.scope_id
      AND extra.group_id = group_row.group_id
      AND extra.domain <> 'work'
  );
