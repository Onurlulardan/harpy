-- +goose Up
-- Migration stream version: 20260729100039.

REVOKE ALL ON FUNCTION platform.billing_installation_assign_v1(uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.billing_installation_assign_v1(uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION platform.billing_installation_assign_v1(uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text) FROM harpy_v2_runtime;
REVOKE ALL ON FUNCTION platform.billing_installation_assign_v2(uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text,text) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.billing_installation_assign_v2(uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text,text) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION platform.billing_installation_assign_v2(uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text,text) FROM harpy_v2_runtime;
REVOKE ALL ON FUNCTION platform.billing_installation_open_usage_v1(uuid,uuid,text) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.billing_installation_open_usage_v1(uuid,uuid,text) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION platform.billing_installation_open_usage_v1(uuid,uuid,text) FROM harpy_v2_runtime;
REVOKE ALL ON FUNCTION platform.billing_reconcile_current_global_authority_v1() FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.billing_reconcile_current_global_authority_v1() FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION platform.billing_reconcile_current_global_authority_v1() FROM harpy_v2_runtime;
REVOKE ALL ON FUNCTION platform.billing_reconcile_current_tenant_authorities_v1() FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.billing_reconcile_current_tenant_authorities_v1() FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION platform.billing_reconcile_current_tenant_authorities_v1() FROM harpy_v2_runtime;
REVOKE ALL ON FUNCTION platform.billing_register_global_authority_v1(uuid,bigint,bigint,timestamptz) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.billing_register_global_authority_v1(uuid,bigint,bigint,timestamptz) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION platform.billing_register_global_authority_v1(uuid,bigint,bigint,timestamptz) FROM harpy_v2_runtime;
REVOKE ALL ON FUNCTION platform.billing_register_tenant_authority_v1(uuid,uuid,bigint,bigint,timestamptz) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.billing_register_tenant_authority_v1(uuid,uuid,bigint,bigint,timestamptz) FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION platform.billing_register_tenant_authority_v1(uuid,uuid,bigint,bigint,timestamptz) FROM harpy_v2_runtime;

WITH retired(signature) AS (
    VALUES
        ('platform.billing_installation_assign_v1(uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text)'),
        ('platform.billing_installation_assign_v2(uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text,text)'),
        ('platform.billing_installation_open_usage_v1(uuid,uuid,text)'),
        ('platform.billing_reconcile_current_global_authority_v1()'),
        ('platform.billing_reconcile_current_tenant_authorities_v1()'),
        ('platform.billing_register_global_authority_v1(uuid,bigint,bigint,timestamptz)'),
        ('platform.billing_register_tenant_authority_v1(uuid,uuid,bigint,bigint,timestamptz)')
), retained(signature) AS (
    VALUES
        ('platform.billing_installation_assign_v3(uuid,uuid,uuid,uuid,uuid,text,uuid,bigint,text,timestamptz,bigint,text,text,text)'),
        ('platform.billing_installation_open_usage_v2(uuid,uuid,text)')
), internal(signature) AS (
    VALUES
        ('platform.billing_installation_id_v1(uuid,text)'),
        ('platform.billing_reconcile_current_global_authority_v2()'),
        ('platform.billing_reconcile_current_tenant_authorities_v2()')
), required_tables(relation_name) AS (
    VALUES ('platform.billing_installation_assignment_contexts')
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
