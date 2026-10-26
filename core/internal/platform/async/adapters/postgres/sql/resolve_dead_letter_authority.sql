SELECT
    group_row.scope_kind,
    group_row.scope_id,
    group_row.group_id,
    group_row.generation,
    domain_row.domain,
    domain_row.generation
FROM platform.cutover_group_members AS member
JOIN platform.cutover_groups AS group_row
  ON group_row.scope_kind = member.scope_kind
 AND group_row.scope_id = member.scope_id
 AND group_row.group_id = member.group_id
JOIN platform.domain_cutovers AS domain_row
  ON domain_row.scope_kind = member.scope_kind
 AND domain_row.scope_id = member.scope_id
 AND domain_row.domain = member.domain
WHERE member.scope_kind = $1
  AND member.scope_id = $2
  AND member.domain = $3
  AND group_row.state = 'v2'
  AND group_row.writer = 'v2'
  AND domain_row.state = 'v2'
  AND domain_row.writer = 'v2'
ORDER BY group_row.group_id COLLATE "C"
;
