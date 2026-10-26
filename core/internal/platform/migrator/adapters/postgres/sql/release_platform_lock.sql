SELECT pg_catalog.pg_advisory_unlock(
    pg_catalog.hashtextextended($1::text, 0::bigint)
)
FROM pg_catalog.pg_database AS database_row
WHERE database_row.datname = pg_catalog.current_database();
