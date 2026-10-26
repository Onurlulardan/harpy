SELECT pg_catalog.count(*)::bigint
FROM pg_catalog.pg_index AS index_catalog
JOIN pg_catalog.pg_class AS relation
  ON relation.oid = index_catalog.indrelid
JOIN pg_catalog.pg_namespace AS namespace
  ON namespace.oid = relation.relnamespace
WHERE namespace.nspname = $1
  AND relation.relname = $2;
