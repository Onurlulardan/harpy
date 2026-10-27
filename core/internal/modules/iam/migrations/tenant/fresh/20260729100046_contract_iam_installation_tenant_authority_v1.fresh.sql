-- +goose Up
-- Migration stream version: 20260729100046.

REVOKE ALL ON FUNCTION iam_installation_initialize_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,text,text,text,text,uuid,uuid,
    uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION iam_installation_initialize_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,text,text,text,text,uuid,uuid,
    uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid
) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION iam_installation_initialize_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,text,text,text,text,uuid,uuid,
    uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid
) FROM harpy_v2_runtime;
REVOKE ALL ON FUNCTION iam_installation_credential_initialize_v1(
    name,uuid,uuid,uuid,uuid,bigint,bytea,bytea
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION iam_installation_credential_initialize_v1(
    name,uuid,uuid,uuid,uuid,bigint,bytea,bytea
) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION iam_installation_credential_initialize_v1(
    name,uuid,uuid,uuid,uuid,bigint,bytea,bytea
) FROM harpy_v2_runtime;
REVOKE ALL ON FUNCTION iam_initial_admin_setup_finalize_v1(
    name,uuid,uuid,text
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION iam_initial_admin_setup_finalize_v1(
    name,uuid,uuid,text
) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION iam_initial_admin_setup_finalize_v1(
    name,uuid,uuid,text
) FROM harpy_v2_runtime;

WITH retired(signature) AS (
    VALUES
        ('iam_installation_initialize_v1(uuid,uuid,uuid,uuid,uuid,uuid,text,text,text,text,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid)'),
        ('iam_installation_credential_initialize_v1(name,uuid,uuid,uuid,uuid,bigint,bytea,bytea)'),
        ('iam_initial_admin_setup_finalize_v1(name,uuid,uuid,text)')
), retained(signature) AS (
    VALUES
        ('iam_installation_initialize_v2(uuid,uuid,uuid,uuid,uuid,uuid,text,text,text,text,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid)'),
        ('iam_installation_credential_initialize_v2(name,uuid,uuid,uuid,uuid,uuid,bigint,bytea,bytea)'),
        ('iam_initial_admin_setup_finalize_v2(name,uuid,uuid,uuid,text)')
), internal(signature) AS (
    VALUES
        ('platform.tenancy_activation_step_context_assert_v1(uuid,text,uuid,uuid,name,text)'),
        ('platform.tenancy_activation_step_consume_v1(uuid,text)')
), roles(role_name) AS (
    VALUES ('harpy_runtime'), ('harpy_v2_runtime')
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
              pg_catalog.to_regrole('harpy_owner')::oid
           OR NOT routine.prosecdef
           OR NOT COALESCE(routine.proconfig, ARRAY[]::text[])
              @> ARRAY[
                  'search_path=' ||
                  pg_catalog.current_setting('search_path')
              ]::text[]
           OR NOT pg_catalog.has_function_privilege(
               'harpy_runtime', routine_oid, 'EXECUTE'
           )
           OR pg_catalog.has_function_privilege(
               'harpy_v2_runtime', routine_oid, 'EXECUTE'
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
