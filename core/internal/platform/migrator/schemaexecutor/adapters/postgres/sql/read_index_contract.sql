SELECT
    index_relation.relowner::regrole::text,
    index_catalog.indisvalid,
    index_catalog.indisready,
    index_catalog.indisunique,
    access_method.amname,
    table_relation.relname,
    pg_catalog.array_agg(attribute.attname ORDER BY key.subscript),
	pg_catalog.array_agg(operator_class.opcname ORDER BY key.subscript),
	pg_catalog.array_agg(operator_namespace.nspname ORDER BY key.subscript),
    pg_catalog.bool_and(operator_class.opcdefault),
    pg_catalog.bool_and(index_catalog.indcollation[key.subscript] = attribute.attcollation),
    pg_catalog.bool_and(index_catalog.indoption[key.subscript] = 0),
    index_catalog.indnkeyatts = index_catalog.indnatts,
    NOT index_catalog.indnullsnotdistinct,
    NOT EXISTS (
        SELECT 1
        FROM pg_catalog.pg_constraint AS backing_constraint
        WHERE backing_constraint.conindid = index_relation.oid
    ),
    COALESCE(
        pg_catalog.pg_get_expr(index_catalog.indpred, index_catalog.indrelid),
        ''
    )
FROM pg_catalog.pg_index AS index_catalog
JOIN pg_catalog.pg_class AS index_relation
  ON index_relation.oid = index_catalog.indexrelid
JOIN pg_catalog.pg_class AS table_relation
  ON table_relation.oid = index_catalog.indrelid
JOIN pg_catalog.pg_namespace AS namespace
  ON namespace.oid = index_relation.relnamespace
JOIN pg_catalog.pg_am AS access_method
  ON access_method.oid = index_relation.relam
JOIN LATERAL pg_catalog.generate_subscripts(
    index_catalog.indkey::smallint[], 1
) AS key(subscript) ON TRUE
JOIN pg_catalog.pg_attribute AS attribute
  ON attribute.attrelid = table_relation.oid
 AND attribute.attnum = index_catalog.indkey[key.subscript]
JOIN pg_catalog.pg_opclass AS operator_class
  ON operator_class.oid = index_catalog.indclass[key.subscript]
JOIN pg_catalog.pg_namespace AS operator_namespace
  ON operator_namespace.oid = operator_class.opcnamespace
WHERE namespace.nspname = $1
  AND index_relation.relname = $2
GROUP BY index_relation.relowner, index_catalog.indisvalid,
    index_catalog.indisready, index_catalog.indisunique, access_method.amname,
    table_relation.relname, index_catalog.indpred, index_catalog.indrelid,
    index_catalog.indnkeyatts, index_catalog.indnatts,
    index_catalog.indnullsnotdistinct, index_relation.oid;
