-- +goose Up
-- Migration stream version: 20260729100042.

REVOKE ALL ON FUNCTION platform.platformops_installation_recovery_initialize_v1(
    uuid,uuid,uuid,name,text,bigint,text,text
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION
    platform.platformops_installation_recovery_initialize_v1(
        uuid,uuid,uuid,name,text,bigint,text,text
    ) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION
    platform.platformops_installation_recovery_initialize_v1(
        uuid,uuid,uuid,name,text,bigint,text,text
    ) FROM harpy_v2_runtime;

WITH retired(signature) AS (
    VALUES (
        'platform.platformops_installation_recovery_initialize_v1(uuid,uuid,uuid,name,text,bigint,text,text)'
    )
), retained(signature) AS (
    VALUES (
        'platform.platformops_installation_recovery_initialize_v2(uuid,uuid,uuid,name,text,bigint,text,text)'
    )
), roles(role_name) AS (
    VALUES ('harpy_runtime'), ('harpy_v2_runtime')
), retained_resolved(signature, routine_oid) AS (
    SELECT signature, pg_catalog.to_regprocedure(signature)
    FROM retained
)
SELECT 1 / CASE WHEN
    EXISTS (
        SELECT 1 FROM retired CROSS JOIN roles
        WHERE pg_catalog.to_regprocedure(signature) IS NULL
           OR pg_catalog.has_function_privilege(
               role_name, pg_catalog.to_regprocedure(signature), 'EXECUTE'
           )
    )
    OR EXISTS (
        SELECT 1 FROM retained_resolved
        LEFT JOIN pg_catalog.pg_proc AS routine
          ON routine.oid = retained_resolved.routine_oid
        WHERE retained_resolved.routine_oid IS NULL
           OR routine.proowner IS DISTINCT FROM
              pg_catalog.to_regrole('harpy_owner')::oid
           OR NOT routine.prosecdef
           OR NOT COALESCE(routine.proconfig, ARRAY[]::text[])
              @> ARRAY['search_path=pg_catalog']::text[]
           OR NOT pg_catalog.has_function_privilege(
               'harpy_runtime', routine.oid, 'EXECUTE'
           )
           OR pg_catalog.has_function_privilege(
               'harpy_v2_runtime', routine.oid, 'EXECUTE'
           )
    )
THEN 0 ELSE 1 END;
