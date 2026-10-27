SELECT
    relation.relkind::text,
    relation.relowner::regrole::text
FROM pg_catalog.pg_class AS relation
JOIN pg_catalog.pg_namespace AS namespace
  ON namespace.oid = relation.relnamespace
WHERE namespace.nspname = $1
  AND relation.relname = $2;
