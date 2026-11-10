SELECT
    1::bigint,
    1::bigint,
    $3,
    $4,
    'ready',
    identity.release_id::text,
    1::bigint,
    TRUE
FROM platform.environment_identity AS identity
JOIN pg_catalog.pg_extension AS extension ON extension.extname = 'vector'
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = extension.extnamespace
JOIN platform.extension_registry AS registry
  ON registry.name = extension.extname
 AND registry.schema_name = namespace.nspname
WHERE identity.singleton = TRUE
  AND identity.database_oid = (
      SELECT database_row.oid::bigint
      FROM pg_catalog.pg_database AS database_row
      WHERE database_row.datname = pg_catalog.current_database()
  )
  AND $1::bigint >= 1
  AND $2::bigint >= 1
  AND $5::text <> '';
