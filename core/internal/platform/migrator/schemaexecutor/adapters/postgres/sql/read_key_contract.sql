SELECT
    source_relation.relowner::regrole::text,
    constraint_catalog.contype::text,
    constraint_catalog.convalidated,
    constraint_catalog.confdeltype::text,
    COALESCE(target_namespace.nspname, ''),
    COALESCE(target_relation.relname, ''),
    COALESCE(target_relation.relowner::regrole::text, ''),
    constraint_catalog.confupdtype::text,
    constraint_catalog.confmatchtype::text,
    constraint_catalog.condeferrable,
    constraint_catalog.condeferred,
    ARRAY(
        SELECT attribute.attname
        FROM pg_catalog.unnest(constraint_catalog.conkey)
          WITH ORDINALITY AS key(attnum, ordinality)
        JOIN pg_catalog.pg_attribute AS attribute
          ON attribute.attrelid = source_relation.oid
         AND attribute.attnum = key.attnum
        ORDER BY key.ordinality
    ),
    ARRAY(
        SELECT attribute.attname
        FROM pg_catalog.unnest(constraint_catalog.confkey)
          WITH ORDINALITY AS key(attnum, ordinality)
        JOIN pg_catalog.pg_attribute AS attribute
          ON attribute.attrelid = target_relation.oid
         AND attribute.attnum = key.attnum
        ORDER BY key.ordinality
    )
FROM pg_catalog.pg_constraint AS constraint_catalog
JOIN pg_catalog.pg_class AS source_relation
  ON source_relation.oid = constraint_catalog.conrelid
JOIN pg_catalog.pg_namespace AS source_namespace
  ON source_namespace.oid = source_relation.relnamespace
LEFT JOIN pg_catalog.pg_class AS target_relation
  ON target_relation.oid = constraint_catalog.confrelid
LEFT JOIN pg_catalog.pg_namespace AS target_namespace
  ON target_namespace.oid = target_relation.relnamespace
WHERE source_namespace.nspname = $1
  AND source_relation.relname = $2
  AND constraint_catalog.conname = $3;
