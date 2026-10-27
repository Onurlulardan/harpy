SELECT max(logical_version)
FROM platform.tenant_migration_checkpoints
WHERE tenant_id = $1 AND state = 'succeeded';
