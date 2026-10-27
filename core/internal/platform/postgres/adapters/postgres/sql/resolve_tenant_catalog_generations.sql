SELECT group_row.generation, catalog_row.generation
FROM platform.cutover_groups AS group_row
JOIN platform.cutover_group_members AS catalog_member
  ON catalog_member.scope_kind = group_row.scope_kind
 AND catalog_member.scope_id = group_row.scope_id
 AND catalog_member.group_id = group_row.group_id
 AND catalog_member.domain = 'catalog'
 AND catalog_member.ordinal = 1
JOIN platform.domain_cutovers AS catalog_row
  ON catalog_row.scope_kind = catalog_member.scope_kind
 AND catalog_row.scope_id = catalog_member.scope_id
 AND catalog_row.domain = catalog_member.domain
WHERE group_row.scope_kind = 'tenant'
  AND group_row.scope_id = $1
  AND group_row.group_id = 'catalog'
  AND group_row.state = 'v2' AND group_row.writer = 'v2'
  AND catalog_row.state = 'v2' AND catalog_row.writer = 'v2'
  AND NOT EXISTS (
    SELECT 1
    FROM platform.cutover_group_members AS extra
    WHERE extra.scope_kind = group_row.scope_kind
      AND extra.scope_id = group_row.scope_id
      AND extra.group_id = group_row.group_id
      AND extra.domain <> 'catalog'
  );
