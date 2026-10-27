SELECT state, writer, generation
FROM platform.domain_cutovers
WHERE scope_kind = 'tenant' AND scope_id = $1 AND domain = $2;
