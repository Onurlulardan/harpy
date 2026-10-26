WITH bound_path AS (
    SELECT pg_catalog.current_schemas(false) AS schemas
    FROM pg_catalog.pg_settings
    WHERE name = 'search_path'
)
SELECT CASE
    WHEN pg_catalog.cardinality(schemas) = 2
     AND schemas[1]::text = 'pg_catalog'
     AND schemas[2]::text <> 'pg_catalog'
    THEN schemas[2]::text
    ELSE ''
END
FROM bound_path
