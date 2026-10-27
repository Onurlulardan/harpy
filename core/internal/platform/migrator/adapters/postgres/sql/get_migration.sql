SELECT
    applied_at,
    is_applied
FROM platform.platform_migration_history
WHERE version = $1;
