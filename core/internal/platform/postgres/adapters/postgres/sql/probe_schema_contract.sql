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
WHERE identity.singleton = TRUE
  AND identity.database_origin = 'fresh'
  AND identity.database_oid = (
      SELECT database_row.oid::bigint
      FROM pg_catalog.pg_database AS database_row
      WHERE database_row.datname = pg_catalog.current_database()
  );
