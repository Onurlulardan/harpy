-- +goose Up
-- The immutable tenant search release already owns database generation 73.
-- This forward repair therefore advances the platform release frontier to 74.
SELECT 1 / CASE WHEN EXISTS (
    SELECT 1
    FROM pg_catalog.pg_proc AS routine
    JOIN pg_catalog.pg_namespace AS namespace_record
      ON namespace_record.oid = routine.pronamespace
    WHERE namespace_record.nspname = 'pg_catalog'
      AND routine.proname = 'normalize'
)
AND EXISTS (
    SELECT 1
    FROM pg_catalog.pg_proc AS routine
    JOIN pg_catalog.pg_namespace AS namespace_record
      ON namespace_record.oid = routine.pronamespace
    WHERE namespace_record.nspname = 'pg_catalog'
      AND routine.proname = 'casefold'
)
AND EXISTS (
    SELECT 1
    FROM pg_catalog.pg_collation AS collation_record
    JOIN pg_catalog.pg_namespace AS namespace_record
      ON namespace_record.oid = collation_record.collnamespace
    WHERE namespace_record.nspname = 'pg_catalog'
      AND collation_record.collname = 'pg_unicode_fast'
)
AND (
    SELECT NOT role_record.rolsuper
       AND NOT role_record.rolinherit
       AND NOT role_record.rolcreaterole
       AND NOT role_record.rolcreatedb
       AND NOT role_record.rolcanlogin
       AND NOT role_record.rolbypassrls
    FROM pg_catalog.pg_roles AS role_record
    WHERE role_record.rolname = 'harpy_runtime'
) THEN 1 ELSE 0 END AS iam_rest_query_search_release_frontier_assertion;
