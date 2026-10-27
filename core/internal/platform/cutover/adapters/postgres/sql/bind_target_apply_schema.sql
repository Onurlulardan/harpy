WITH target AS MATERIALIZED (
    SELECT namespace.oid::bigint AS schema_oid,
           relation.oid::bigint AS relation_oid,
           namespace.nspname
    FROM pg_catalog.pg_namespace AS namespace
    JOIN pg_catalog.pg_class AS relation
      ON relation.relnamespace = namespace.oid
    WHERE namespace.oid = $1::oid
      AND namespace.nspname = $2::text
      AND relation.oid = $3::oid
      AND relation.relname = $4::text
      AND relation.relkind IN ('r', 'p')
), expected_procedure AS MATERIALIZED (
    SELECT procedure.oid::bigint AS procedure_oid,
           namespace.nspname AS schema_name
    FROM target
    JOIN pg_catalog.pg_proc AS procedure ON procedure.oid = $6::oid
    JOIN pg_catalog.pg_namespace AS namespace
      ON namespace.oid = procedure.pronamespace
    WHERE procedure.prokind = 'f'
      AND procedure.pronamespace = target.schema_oid::oid
      AND pg_catalog.to_regprocedure($5::text)::oid = procedure.oid
), bound AS MATERIALIZED (
    SELECT target.schema_oid, target.relation_oid,
           expected_procedure.procedure_oid,
           pg_catalog.set_config(
               'search_path',
               pg_catalog.format('%I,pg_catalog', expected_procedure.schema_name),
               true
           ) AS search_path
    FROM target
    CROSS JOIN expected_procedure
)
SELECT bound.schema_oid, bound.relation_oid,
       bound.procedure_oid
FROM bound
