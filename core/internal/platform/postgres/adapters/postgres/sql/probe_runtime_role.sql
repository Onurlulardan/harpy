SELECT pg_catalog.set_config('role', 'harpy_worker', true)
FROM pg_catalog.pg_roles
WHERE rolname = 'harpy_worker';
