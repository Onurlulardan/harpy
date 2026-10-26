SELECT group_row.generation, documents_row.generation
FROM platform.cutover_groups AS group_row
JOIN platform.cutover_group_members AS documents_member
  ON documents_member.scope_kind = group_row.scope_kind
 AND documents_member.scope_id = group_row.scope_id
 AND documents_member.group_id = group_row.group_id
 AND documents_member.domain = 'documents'
 AND documents_member.ordinal = 1
JOIN platform.domain_cutovers AS documents_row
  ON documents_row.scope_kind = documents_member.scope_kind
 AND documents_row.scope_id = documents_member.scope_id
 AND documents_row.domain = documents_member.domain
WHERE group_row.scope_kind = 'tenant'
  AND group_row.scope_id = $1
  AND group_row.group_id = 'documents'
  AND group_row.state = 'v2' AND group_row.writer = 'v2'
  AND documents_row.state = 'v2' AND documents_row.writer = 'v2'
  AND NOT EXISTS (
    SELECT 1
    FROM platform.cutover_group_members AS extra
    WHERE extra.scope_kind = group_row.scope_kind
      AND extra.scope_id = group_row.scope_id
      AND extra.group_id = group_row.group_id
      AND extra.domain <> 'documents'
  );
