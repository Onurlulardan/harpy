SELECT updated_at, state = 'succeeded'
FROM platform.tenant_migration_checkpoints
WHERE tenant_id = $1 AND logical_version = $2;
