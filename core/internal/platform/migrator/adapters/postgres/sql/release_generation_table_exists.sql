SELECT pg_catalog.count(*) = 1
FROM pg_catalog.pg_class AS relation
JOIN pg_catalog.pg_namespace AS namespace
  ON namespace.oid = relation.relnamespace
WHERE namespace.nspname = 'platform'
  AND relation.relname = 'release_generations'
  AND relation.relkind = 'r';
