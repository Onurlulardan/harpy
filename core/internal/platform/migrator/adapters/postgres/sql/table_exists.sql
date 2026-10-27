SELECT pg_catalog.count(*) = 1
FROM pg_catalog.pg_class AS relation
INNER JOIN pg_catalog.pg_namespace AS namespace
    ON namespace.oid = relation.relnamespace
WHERE namespace.nspname = 'platform'
  AND relation.relname = 'platform_migration_history'
  AND relation.relkind = 'r';
