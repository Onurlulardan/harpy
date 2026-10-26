SELECT group_row.generation,
       billing_row.generation,
       forms_row.generation
FROM platform.cutover_groups AS group_row
JOIN platform.cutover_group_members AS billing_member
  ON billing_member.scope_kind = group_row.scope_kind
 AND billing_member.scope_id = group_row.scope_id
 AND billing_member.group_id = group_row.group_id
 AND billing_member.domain = 'billing'
JOIN platform.cutover_group_members AS forms_member
  ON forms_member.scope_kind = group_row.scope_kind
 AND forms_member.scope_id = group_row.scope_id
 AND forms_member.group_id = group_row.group_id
 AND forms_member.domain = 'forms'
JOIN platform.domain_cutovers AS billing_row
  ON billing_row.scope_kind = billing_member.scope_kind
 AND billing_row.scope_id = billing_member.scope_id
 AND billing_row.domain = billing_member.domain
JOIN platform.domain_cutovers AS forms_row
  ON forms_row.scope_kind = forms_member.scope_kind
 AND forms_row.scope_id = forms_member.scope_id
 AND forms_row.domain = forms_member.domain
WHERE group_row.scope_kind = 'tenant'
  AND group_row.scope_id = $1
  AND group_row.group_id = 'billing_forms'
  AND group_row.state = 'v2' AND group_row.writer = 'v2'
  AND billing_row.state = 'v2' AND billing_row.writer = 'v2'
  AND forms_row.state = 'v2' AND forms_row.writer = 'v2'
  AND NOT EXISTS (
    SELECT 1
    FROM platform.cutover_group_members AS extra
    WHERE extra.scope_kind = group_row.scope_kind
      AND extra.scope_id = group_row.scope_id
      AND extra.group_id = group_row.group_id
      AND extra.domain NOT IN ('billing', 'forms')
  );
