SELECT
    pg_catalog.format_type(attribute.atttypid, attribute.atttypmod),
    attribute.attnotnull,
    COALESCE(
        pg_catalog.pg_get_expr(attribute_default.adbin, attribute_default.adrelid),
        ''
    ),
    attribute.attcollation = type_catalog.typcollation
FROM pg_catalog.pg_attribute AS attribute
JOIN pg_catalog.pg_class AS relation
  ON relation.oid = attribute.attrelid
JOIN pg_catalog.pg_namespace AS namespace
  ON namespace.oid = relation.relnamespace
JOIN pg_catalog.pg_type AS type_catalog
  ON type_catalog.oid = attribute.atttypid
LEFT JOIN pg_catalog.pg_attrdef AS attribute_default
  ON attribute_default.adrelid = attribute.attrelid
 AND attribute_default.adnum = attribute.attnum
WHERE namespace.nspname = $1
  AND relation.relname = $2
  AND attribute.attname = $3
  AND attribute.attnum > 0
  AND NOT attribute.attisdropped;
