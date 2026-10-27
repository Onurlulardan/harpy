SELECT
    relation.relowner::regrole::text,
    constraint_catalog.contype::text,
    constraint_catalog.convalidated,
    pg_catalog.pg_get_constraintdef(constraint_catalog.oid, false)
FROM pg_catalog.pg_constraint AS constraint_catalog
JOIN pg_catalog.pg_class AS relation
  ON relation.oid = constraint_catalog.conrelid
JOIN pg_catalog.pg_namespace AS namespace
  ON namespace.oid = relation.relnamespace
WHERE namespace.nspname = $1
  AND relation.relname = $2
  AND constraint_catalog.conname = $3;
