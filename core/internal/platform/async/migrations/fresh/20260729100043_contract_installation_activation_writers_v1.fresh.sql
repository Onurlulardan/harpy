-- +goose Up
-- Migration stream version: 20260729100043.

REVOKE ALL ON FUNCTION platform.platform_async_initial_admin_setup_enqueue_v1(
    uuid,uuid,uuid,text,name,text
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION
    platform.platform_async_initial_admin_setup_enqueue_v1(
        uuid,uuid,uuid,text,name,text
    ) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION
    platform.platform_async_initial_admin_setup_enqueue_v1(
        uuid,uuid,uuid,text,name,text
    ) FROM forgecrud_v2_runtime;
REVOKE ALL ON FUNCTION
    platform.platform_async_installation_activation_event_v1(
        uuid,uuid,uuid,uuid,name,text
    ) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION
    platform.platform_async_installation_activation_event_v1(
        uuid,uuid,uuid,uuid,name,text
    ) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION
    platform.platform_async_installation_activation_event_v1(
        uuid,uuid,uuid,uuid,name,text
    ) FROM forgecrud_v2_runtime;

WITH retired(signature) AS (
    VALUES
        ('platform.platform_async_initial_admin_setup_enqueue_v1(uuid,uuid,uuid,text,name,text)'),
        ('platform.platform_async_installation_activation_event_v1(uuid,uuid,uuid,uuid,name,text)')
), retained(signature) AS (
    VALUES
        ('platform.platform_async_initial_admin_setup_enqueue_v2(uuid,uuid,uuid,uuid,text,name,text)'),
        ('platform.platform_async_installation_activation_event_v2(uuid,uuid,uuid,uuid,name,text)')
), internal(signature) AS (
    VALUES (
        'platform.platform_async_installation_activation_evidence_assert_v1(uuid,uuid)'
    )
), roles(role_name) AS (
    VALUES ('forgecrud_runtime'), ('forgecrud_v2_runtime')
), retained_resolved(signature, routine_oid) AS (
    SELECT signature, pg_catalog.to_regprocedure(signature)
    FROM retained
), receipt_table(table_oid) AS (
    SELECT pg_catalog.to_regclass(
        'platform.platform_async_installation_activation_receipts'
    )
), table_privileges(privilege_name) AS (
    VALUES
        ('SELECT'), ('INSERT'), ('UPDATE'), ('DELETE'),
        ('TRUNCATE'), ('REFERENCES'), ('TRIGGER')
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
              pg_catalog.to_regrole('forgecrud_owner')::oid
           OR NOT routine.prosecdef
           OR NOT COALESCE(routine.proconfig, ARRAY[]::text[])
              @> ARRAY['search_path=pg_catalog']::text[]
           OR NOT pg_catalog.has_function_privilege(
               'forgecrud_runtime', routine.oid, 'EXECUTE'
           )
           OR pg_catalog.has_function_privilege(
               'forgecrud_v2_runtime', routine.oid, 'EXECUTE'
           )
    )
    OR EXISTS (
        SELECT 1 FROM internal CROSS JOIN roles
        WHERE pg_catalog.to_regprocedure(signature) IS NULL
           OR pg_catalog.has_function_privilege(
               role_name, pg_catalog.to_regprocedure(signature), 'EXECUTE'
           )
    )
    OR EXISTS (
        SELECT 1
        FROM roles
        CROSS JOIN receipt_table
        CROSS JOIN table_privileges
        WHERE receipt_table.table_oid IS NULL
           OR pg_catalog.has_table_privilege(
               role_name, receipt_table.table_oid, privilege_name
           )
    )
THEN 0 ELSE 1 END;
