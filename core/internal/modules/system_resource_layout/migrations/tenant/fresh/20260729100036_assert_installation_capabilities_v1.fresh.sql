-- +goose Up
-- Migration stream version: 20260729100036.

WITH public_routines(signature) AS (
    VALUES
        ('system_resource_layout_installation_assert_v2(uuid,uuid,name)'),
        ('system_resource_layout_installation_record_v2(uuid,uuid,uuid,uuid,uuid,timestamptz,name)')
), internal_routines(signature) AS (
    VALUES
        ('platform.tenancy_activation_step_context_assert_v1(uuid,text,uuid,uuid,name,text)'),
        ('platform.tenancy_activation_step_consume_v1(uuid,text)')
), legacy_routines(signature) AS (
    VALUES
        ('system_resource_layout_installation_record_v1(uuid,uuid,uuid,uuid,uuid,timestamptz)')
), public_resolved AS (
    SELECT signature, pg_catalog.to_regprocedure(signature) AS routine_oid
    FROM public_routines
), internal_resolved AS (
    SELECT signature, pg_catalog.to_regprocedure(signature) AS routine_oid
    FROM internal_routines
), legacy_resolved AS (
    SELECT signature, pg_catalog.to_regprocedure(signature) AS routine_oid
    FROM legacy_routines
)
SELECT 1 / CASE WHEN
    EXISTS (
        SELECT 1
        FROM public_resolved
        LEFT JOIN pg_catalog.pg_proc AS routine
          ON routine.oid = public_resolved.routine_oid
        WHERE routine_oid IS NULL
           OR routine.proowner IS DISTINCT FROM
              pg_catalog.to_regrole('forgecrud_owner')::oid
           OR NOT routine.prosecdef
           OR NOT (
               COALESCE(routine.proconfig, ARRAY[]::text[])
               @> ARRAY[
                   'search_path=' ||
                   pg_catalog.current_setting('search_path')
               ]::text[]
           )
           OR NOT pg_catalog.has_function_privilege(
               'forgecrud_runtime', routine_oid, 'EXECUTE'
           )
           OR pg_catalog.has_function_privilege(
               'forgecrud_v2_runtime', routine_oid, 'EXECUTE'
           )
    )
    OR EXISTS (
        SELECT 1
        FROM internal_resolved
        WHERE routine_oid IS NULL
           OR pg_catalog.has_function_privilege(
               'forgecrud_runtime', routine_oid, 'EXECUTE'
           )
           OR pg_catalog.has_function_privilege(
               'forgecrud_v2_runtime', routine_oid, 'EXECUTE'
           )
    )
    OR EXISTS (
        SELECT 1
        FROM legacy_resolved
        WHERE routine_oid IS NULL
           OR NOT pg_catalog.has_function_privilege(
               'forgecrud_runtime', routine_oid, 'EXECUTE'
           )
    )
THEN 0 ELSE 1 END;
