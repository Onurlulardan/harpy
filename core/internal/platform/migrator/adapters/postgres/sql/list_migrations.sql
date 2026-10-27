SELECT
    version,
    is_applied
FROM platform.platform_migration_history
ORDER BY version DESC;
