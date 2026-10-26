SELECT pg_catalog.current_schemas(false)
FROM pg_catalog.pg_settings
WHERE name = 'search_path';
