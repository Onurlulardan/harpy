SELECT pg_catalog.set_config('role', 'harpy_runtime', true)
FROM pg_catalog.pg_roles
WHERE rolname = 'harpy_runtime';
