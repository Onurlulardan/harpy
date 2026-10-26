SELECT state, writer, generation
FROM platform.domain_cutovers
WHERE scope_kind = 'global' AND scope_id = $1 AND domain = 'billing';
