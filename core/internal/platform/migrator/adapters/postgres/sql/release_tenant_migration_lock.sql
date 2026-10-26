SELECT $1::text,
       pg_catalog.pg_advisory_unlock(
           pg_catalog.hashtextextended('harpy:v2:migrate:tenant:' || $1, 0)
       )
FROM pg_catalog.pg_database AS database_row
WHERE database_row.datname = pg_catalog.current_database();
