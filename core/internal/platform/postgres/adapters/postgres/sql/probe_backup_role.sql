SELECT pg_catalog.set_config('role', 'harpy_backup_reader', true)
FROM pg_catalog.pg_roles
WHERE rolname = 'harpy_backup_reader';
