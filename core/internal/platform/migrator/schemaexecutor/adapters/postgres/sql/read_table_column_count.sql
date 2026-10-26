SELECT pg_catalog.count(*)::bigint
FROM pg_catalog.pg_attribute AS attribute
JOIN pg_catalog.pg_class AS relation
  ON relation.oid = attribute.attrelid
JOIN pg_catalog.pg_namespace AS namespace
  ON namespace.oid = relation.relnamespace
WHERE namespace.nspname = $1
  AND relation.relname = $2
  AND attribute.attnum > 0
  AND NOT attribute.attisdropped;
