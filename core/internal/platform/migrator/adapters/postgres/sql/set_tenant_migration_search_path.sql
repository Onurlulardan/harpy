SELECT pg_catalog.set_config('search_path', $1, false)
FROM pg_catalog.pg_database AS database_row
WHERE database_row.datname = pg_catalog.current_database();
