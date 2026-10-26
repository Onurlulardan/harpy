SELECT pg_catalog.pg_backend_pid(),
       pg_catalog.current_setting('search_path'),
       pg_catalog.current_schema()
FROM pg_catalog.pg_database AS database_row
WHERE database_row.datname = pg_catalog.current_database();
