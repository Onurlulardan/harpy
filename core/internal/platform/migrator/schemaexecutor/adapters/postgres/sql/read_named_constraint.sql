SELECT
    relation.relowner::regrole::text,
    constraint_catalog.convalidated
FROM pg_catalog.pg_constraint AS constraint_catalog
JOIN pg_catalog.pg_class AS relation
  ON relation.oid = constraint_catalog.conrelid
JOIN pg_catalog.pg_namespace AS namespace
  ON namespace.oid = relation.relnamespace
WHERE namespace.nspname = $1
  AND constraint_catalog.conname = $2
  AND relation.relname = $3;
