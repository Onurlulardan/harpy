SELECT group_row.generation, workflow_row.generation
FROM platform.cutover_groups AS group_row
JOIN platform.cutover_group_members AS workflow_member
  ON workflow_member.scope_kind = group_row.scope_kind
 AND workflow_member.scope_id = group_row.scope_id
 AND workflow_member.group_id = group_row.group_id
 AND workflow_member.domain = 'workflow'
 AND workflow_member.ordinal = 1
JOIN platform.domain_cutovers AS workflow_row
  ON workflow_row.scope_kind = workflow_member.scope_kind
 AND workflow_row.scope_id = workflow_member.scope_id
 AND workflow_row.domain = workflow_member.domain
WHERE group_row.scope_kind = 'tenant'
  AND group_row.scope_id = $1
  AND group_row.group_id = 'workflow'
  AND group_row.state = 'v2' AND group_row.writer = 'v2'
  AND workflow_row.state = 'v2' AND workflow_row.writer = 'v2'
  AND NOT EXISTS (
    SELECT 1
    FROM platform.cutover_group_members AS extra
    WHERE extra.scope_kind = group_row.scope_kind
      AND extra.scope_id = group_row.scope_id
      AND extra.group_id = group_row.group_id
      AND extra.domain <> 'workflow'
  );
