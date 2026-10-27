SELECT pg_catalog.to_regclass('platform.tenant_migration_checkpoints') IS NOT NULL
FROM pg_catalog.pg_database AS database_row
WHERE database_row.datname = pg_catalog.current_database();
