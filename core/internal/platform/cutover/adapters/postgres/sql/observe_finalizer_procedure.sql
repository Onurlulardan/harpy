SELECT
    'cutover_authority.finalize_operation_v1(uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea)'::regprocedure::oid::bigint,
    pg_catalog.format(
        '%I.%I(%s)', namespace.nspname, procedure_row.proname,
        pg_catalog.replace(pg_catalog.oidvectortypes(procedure_row.proargtypes), ' ', '')
    )
FROM pg_catalog.pg_proc AS procedure_row
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = procedure_row.pronamespace
WHERE procedure_row.oid =
    'cutover_authority.finalize_operation_v1(uuid,uuid,uuid,text,bytea,text,bytea,text,bytea,text,bytea,text,bytea,text,text,bytea)'::regprocedure::oid
