-- +goose Up
-- Migration stream version: 20260729100047.

REVOKE ALL ON FUNCTION system_resource_layout_installation_record_v1(
    uuid,uuid,uuid,uuid,uuid,timestamptz
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION system_resource_layout_installation_record_v1(
    uuid,uuid,uuid,uuid,uuid,timestamptz
) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION system_resource_layout_installation_record_v1(
    uuid,uuid,uuid,uuid,uuid,timestamptz
) FROM forgecrud_v2_runtime;

WITH retired(signature) AS (
    VALUES (
        'system_resource_layout_installation_record_v1(uuid,uuid,uuid,uuid,uuid,timestamptz)'
    )
), retained(signature) AS (
    VALUES
        ('system_resource_layout_installation_assert_v2(uuid,uuid,name)'),
        ('system_resource_layout_installation_record_v2(uuid,uuid,uuid,uuid,uuid,timestamptz,name)')
), internal(signature) AS (
    VALUES
        ('system_resource_layout_installation_baseline_guard_v1()'),
        ('platform.tenancy_activation_step_context_assert_v1(uuid,text,uuid,uuid,name,text)'),
        ('platform.tenancy_activation_step_consume_v1(uuid,text)')
), roles(role_name) AS (
    VALUES ('forgecrud_runtime'), ('forgecrud_v2_runtime')
), retired_resolved AS (
    SELECT signature, pg_catalog.to_regprocedure(signature) AS routine_oid
    FROM retired
), retained_resolved AS (
    SELECT signature, pg_catalog.to_regprocedure(signature) AS routine_oid
    FROM retained
), internal_resolved AS (
    SELECT signature, pg_catalog.to_regprocedure(signature) AS routine_oid
    FROM internal
)
SELECT 1 / CASE WHEN
    EXISTS (
        SELECT 1 FROM retired_resolved CROSS JOIN roles
        WHERE routine_oid IS NULL
           OR pg_catalog.has_function_privilege(
               role_name, routine_oid, 'EXECUTE'
           )
    )
    OR EXISTS (
        SELECT 1
        FROM retained_resolved
        LEFT JOIN pg_catalog.pg_proc AS routine
          ON routine.oid = retained_resolved.routine_oid
        WHERE routine_oid IS NULL
           OR routine.proowner IS DISTINCT FROM
              pg_catalog.to_regrole('forgecrud_owner')::oid
           OR NOT routine.prosecdef
           OR NOT COALESCE(routine.proconfig, ARRAY[]::text[])
              @> ARRAY[
                  'search_path=' ||
                  pg_catalog.current_setting('search_path')
              ]::text[]
           OR NOT pg_catalog.has_function_privilege(
               'forgecrud_runtime', routine_oid, 'EXECUTE'
           )
           OR pg_catalog.has_function_privilege(
               'forgecrud_v2_runtime', routine_oid, 'EXECUTE'
           )
    )
    OR EXISTS (
        SELECT 1 FROM internal_resolved CROSS JOIN roles
        WHERE routine_oid IS NULL
           OR pg_catalog.has_function_privilege(
               role_name, routine_oid, 'EXECUTE'
           )
    )
THEN 0 ELSE 1 END;
