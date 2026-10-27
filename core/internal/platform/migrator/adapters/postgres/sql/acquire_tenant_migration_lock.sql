WITH locked AS MATERIALIZED (
    SELECT pg_catalog.pg_advisory_lock(
        pg_catalog.hashtextextended('harpy:v2:migrate:tenant:' || $1, 0)
    ) AS acquired
    FROM pg_catalog.pg_database AS database_row
    WHERE database_row.datname = pg_catalog.current_database()
)
SELECT $1::text FROM locked;
