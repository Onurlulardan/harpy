SELECT group_row.generation, audit_row.generation
FROM platform.cutover_groups AS group_row
JOIN platform.cutover_group_members AS audit_member
  ON audit_member.scope_kind = group_row.scope_kind
 AND audit_member.scope_id = group_row.scope_id
 AND audit_member.group_id = group_row.group_id
 AND audit_member.domain = 'audit'
 AND audit_member.ordinal = 1
JOIN platform.domain_cutovers AS audit_row
  ON audit_row.scope_kind = audit_member.scope_kind
 AND audit_row.scope_id = audit_member.scope_id
 AND audit_row.domain = audit_member.domain
WHERE group_row.scope_kind = 'tenant'
  AND group_row.scope_id = $1
  AND group_row.group_id = 'audit'
  AND group_row.state = 'v2' AND group_row.writer = 'v2'
  AND audit_row.state = 'v2' AND audit_row.writer = 'v2'
  AND NOT EXISTS (
    SELECT 1
    FROM platform.cutover_group_members AS extra
    WHERE extra.scope_kind = group_row.scope_kind
      AND extra.scope_id = group_row.scope_id
      AND extra.group_id = group_row.group_id
      AND extra.domain <> 'audit'
  );
