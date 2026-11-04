SELECT pg_catalog.set_config('role', 'harpy_api', true)
FROM pg_catalog.pg_roles
WHERE rolname = 'harpy_api';
