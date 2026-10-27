-- +goose Up
-- Migration stream version: 20260729100037.

REVOKE ALL ON FUNCTION platform.tenancy_onboarding_accept_v1(uuid,uuid,text,text,text,text,text,text,uuid,name,text,uuid,uuid,uuid,uuid,uuid,bigint,text,timestamptz,bigint,text,bigint,text,uuid,text,uuid,bigint,text,text,text,timestamptz) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.tenancy_onboarding_accept_v1(uuid,uuid,text,text,text,text,text,text,uuid,name,text,uuid,uuid,uuid,uuid,uuid,bigint,text,timestamptz,bigint,text,bigint,text,uuid,text,uuid,bigint,text,text,text,timestamptz) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION platform.tenancy_onboarding_accept_v1(uuid,uuid,text,text,text,text,text,text,uuid,name,text,uuid,uuid,uuid,uuid,uuid,bigint,text,timestamptz,bigint,text,bigint,text,uuid,text,uuid,bigint,text,text,text,timestamptz) FROM harpy_v2_runtime;
REVOKE ALL ON FUNCTION platform.tenancy_onboarding_activate_v1(uuid,text,bigint) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.tenancy_onboarding_activate_v1(uuid,text,bigint) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION platform.tenancy_onboarding_activate_v1(uuid,text,bigint) FROM harpy_v2_runtime;
REVOKE ALL ON FUNCTION platform.tenancy_onboarding_activate_v2(uuid,text,bigint) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.tenancy_onboarding_activate_v2(uuid,text,bigint) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION platform.tenancy_onboarding_activate_v2(uuid,text,bigint) FROM harpy_v2_runtime;
REVOKE ALL ON FUNCTION platform.tenancy_activation_initialize_v1(uuid,uuid,uuid,text,text,uuid,text,text,text) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.tenancy_activation_initialize_v1(uuid,uuid,uuid,text,text,uuid,text,text,text) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION platform.tenancy_activation_initialize_v1(uuid,uuid,uuid,text,text,uuid,text,text,text) FROM harpy_v2_runtime;
REVOKE ALL ON FUNCTION platform.tenancy_activation_initialize_v2(uuid,uuid,uuid,text,text,uuid,text,text,text) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.tenancy_activation_initialize_v2(uuid,uuid,uuid,text,text,uuid,text,text,text) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION platform.tenancy_activation_initialize_v2(uuid,uuid,uuid,text,text,uuid,text,text,text) FROM harpy_v2_runtime;
REVOKE ALL ON FUNCTION platform.tenancy_activation_initialize_v3(uuid,uuid,uuid,text,text,uuid,text,text,text) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.tenancy_activation_initialize_v3(uuid,uuid,uuid,text,text,uuid,text,text,text) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION platform.tenancy_activation_initialize_v3(uuid,uuid,uuid,text,text,uuid,text,text,text) FROM harpy_v2_runtime;
REVOKE ALL ON FUNCTION platform.tenancy_activation_initialize_v4(uuid,uuid,uuid,text,text,uuid,text,text,text,text) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.tenancy_activation_initialize_v4(uuid,uuid,uuid,text,text,uuid,text,text,text,text) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION platform.tenancy_activation_initialize_v4(uuid,uuid,uuid,text,text,uuid,text,text,text,text) FROM harpy_v2_runtime;
REVOKE ALL ON FUNCTION platform.tenancy_onboarding_complete_v1(uuid,uuid,text,text,bigint) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.tenancy_onboarding_complete_v1(uuid,uuid,text,text,bigint) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION platform.tenancy_onboarding_complete_v1(uuid,uuid,text,text,bigint) FROM harpy_v2_runtime;
REVOKE ALL ON FUNCTION platform.tenancy_activation_assert_v1(uuid,uuid,name) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.tenancy_activation_assert_v1(uuid,uuid,name) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION platform.tenancy_activation_assert_v1(uuid,uuid,name) FROM harpy_v2_runtime;

WITH retired(signature) AS (
    VALUES
        ('platform.tenancy_onboarding_accept_v1(uuid,uuid,text,text,text,text,text,text,uuid,name,text,uuid,uuid,uuid,uuid,uuid,bigint,text,timestamptz,bigint,text,bigint,text,uuid,text,uuid,bigint,text,text,text,timestamptz)'),
        ('platform.tenancy_onboarding_activate_v1(uuid,text,bigint)'),
        ('platform.tenancy_onboarding_activate_v2(uuid,text,bigint)'),
        ('platform.tenancy_activation_initialize_v1(uuid,uuid,uuid,text,text,uuid,text,text,text)'),
        ('platform.tenancy_activation_initialize_v2(uuid,uuid,uuid,text,text,uuid,text,text,text)'),
        ('platform.tenancy_activation_initialize_v3(uuid,uuid,uuid,text,text,uuid,text,text,text)'),
        ('platform.tenancy_activation_initialize_v4(uuid,uuid,uuid,text,text,uuid,text,text,text,text)'),
        ('platform.tenancy_onboarding_complete_v1(uuid,uuid,text,text,bigint)'),
        ('platform.tenancy_activation_assert_v1(uuid,uuid,name)')
), retained(signature) AS (
    VALUES
        ('platform.tenancy_onboarding_accept_v2(uuid,uuid,text,text,text,text,text,text,uuid,name,text,uuid,uuid,uuid,uuid,uuid,bigint,text,timestamptz,bigint,text,bigint,text,uuid,text,uuid,bigint,text,text,text,timestamptz)'),
        ('platform.tenancy_onboarding_activate_v3(uuid,text,bigint)'),
        ('platform.tenancy_activation_initialize_v5(uuid,uuid,uuid,text,text,uuid,text,text,text,text)'),
        ('platform.tenancy_onboarding_complete_v2(uuid,uuid,text,text,bigint)'),
        ('platform.tenancy_activation_assert_v2(uuid,uuid,name)'),
        ('platform.tenancy_authority_reconcile_current_v1(uuid,text)')
), internal(signature) AS (
    VALUES
        ('platform.tenancy_activation_command_consume_v1(uuid,text,uuid,uuid,text,text,text,text,uuid,bigint,text,timestamptz,bigint,text)'),
        ('platform.tenancy_authority_reconcile_scope_assert_v1(text)'),
        ('platform.tenancy_activation_capability_commit_guard_v1()'),
        ('platform.tenancy_authority_scope_commit_guard_v1()'),
        ('platform.tenancy_activation_step_assert_v1(uuid,text)'),
        ('platform.tenancy_activation_step_consume_v1(uuid,text)'),
        ('platform.tenancy_activation_step_context_assert_v1(uuid,text,uuid,uuid,name,text)'),
        ('platform.tenancy_activation_step_commit_guard_v1()')
), required_tables(relation_name) AS (
    VALUES
        ('platform.tenancy_activation_command_capabilities'),
        ('platform.tenancy_authority_reconcile_scopes'),
        ('platform.tenancy_activation_step_capabilities')
), retired_resolved AS (
    SELECT signature, pg_catalog.to_regprocedure(signature) AS routine_oid
    FROM retired
), retained_resolved AS (
    SELECT signature, pg_catalog.to_regprocedure(signature) AS routine_oid
    FROM retained
), internal_resolved AS (
    SELECT signature, pg_catalog.to_regprocedure(signature) AS routine_oid
    FROM internal
), table_resolved AS (
    SELECT relation_name, pg_catalog.to_regclass(relation_name) AS relation_oid
    FROM required_tables
), forbidden_roles(role_name) AS (
    VALUES ('harpy_runtime'), ('harpy_v2_runtime')
), forbidden_privileges(privilege_name) AS (
    VALUES
        ('SELECT'), ('INSERT'), ('UPDATE'), ('DELETE'),
        ('TRUNCATE'), ('REFERENCES'), ('TRIGGER')
)
SELECT 1 / CASE WHEN
    EXISTS (
        SELECT 1
        FROM retired_resolved
        WHERE routine_oid IS NULL
           OR pg_catalog.has_function_privilege(
               'harpy_runtime', routine_oid, 'EXECUTE'
           )
           OR pg_catalog.has_function_privilege(
               'harpy_v2_runtime', routine_oid, 'EXECUTE'
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
           OR NOT (
               COALESCE(routine.proconfig, ARRAY[]::text[])
               @> ARRAY['search_path=pg_catalog']::text[]
           )
           OR NOT pg_catalog.has_function_privilege(
               'harpy_runtime', routine_oid, 'EXECUTE'
           )
           OR pg_catalog.has_function_privilege(
               'harpy_v2_runtime', routine_oid, 'EXECUTE'
           )
    )
    OR EXISTS (
        SELECT 1
        FROM internal_resolved
        LEFT JOIN pg_catalog.pg_proc AS routine
          ON routine.oid = internal_resolved.routine_oid
        WHERE routine_oid IS NULL
           OR routine.proowner IS DISTINCT FROM
              pg_catalog.to_regrole('harpy_owner')::oid
           OR NOT routine.prosecdef
           OR NOT (
               COALESCE(routine.proconfig, ARRAY[]::text[])
               @> ARRAY['search_path=pg_catalog']::text[]
           )
           OR pg_catalog.has_function_privilege(
               'harpy_runtime', routine_oid, 'EXECUTE'
           )
           OR pg_catalog.has_function_privilege(
               'harpy_v2_runtime', routine_oid, 'EXECUTE'
           )
    )
    OR EXISTS (
        SELECT 1
        FROM table_resolved
        LEFT JOIN pg_catalog.pg_class AS relation
          ON relation.oid = table_resolved.relation_oid
        WHERE relation_oid IS NULL
           OR relation.relkind IS DISTINCT FROM 'r'
           OR relation.relowner IS DISTINCT FROM
              pg_catalog.to_regrole('harpy_owner')::oid
    )
    OR EXISTS (
        SELECT 1
        FROM table_resolved
        CROSS JOIN forbidden_roles
        CROSS JOIN forbidden_privileges
        WHERE pg_catalog.has_table_privilege(
            role_name, relation_oid, privilege_name
        )
    )
THEN 0 ELSE 1 END;
