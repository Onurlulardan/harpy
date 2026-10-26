SELECT pg_catalog.set_config('search_path', $1, true)
FROM pg_catalog.pg_settings
WHERE name = 'search_path';
