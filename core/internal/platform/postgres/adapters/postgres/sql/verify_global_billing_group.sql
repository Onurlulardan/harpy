SELECT state, writer, generation
FROM platform.cutover_groups
WHERE scope_kind = 'global' AND scope_id = $1 AND group_id = 'billing';
