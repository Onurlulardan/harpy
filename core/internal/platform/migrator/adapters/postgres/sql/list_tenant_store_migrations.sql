SELECT logical_version, true
FROM platform.tenant_migration_checkpoints
WHERE tenant_id = $1 AND state = 'succeeded'
ORDER BY logical_version DESC;
