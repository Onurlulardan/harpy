SELECT domain
FROM platform.cutover_group_members
WHERE scope_kind = 'global' AND scope_id = $1 AND group_id = 'billing'
ORDER BY ordinal, domain COLLATE "C";
