SELECT
    1::bigint,
    1::bigint,
    $3,
    $4,
    'ready',
    identity.release_id::text,
    1::bigint,
    (
        SELECT pg_catalog.count(*) FILTER (
            WHERE history.version_id IN (20261028000000, 20261102000000)
        ) = 2
        FROM platform.goose_db_version AS history
        WHERE history.is_applied
          AND history.version_id <> 0
    )
FROM platform.environment_identity AS identity
WHERE identity.singleton = TRUE
  AND identity.database_origin = 'fresh'
  AND identity.database_oid = (
      SELECT database_row.oid::bigint
      FROM pg_catalog.pg_database AS database_row
      WHERE database_row.datname = pg_catalog.current_database()
  );
