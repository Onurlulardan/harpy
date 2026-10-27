-- +goose Up
-- Migration stream version: 20260729100040.

REVOKE ALL ON FUNCTION
    platform.analytics_reconcile_current_tenant_authorities_v1() FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION
    platform.analytics_reconcile_current_tenant_authorities_v1()
    FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION
    platform.analytics_reconcile_current_tenant_authorities_v1()
    FROM forgecrud_v2_runtime;
REVOKE ALL ON FUNCTION platform.analytics_register_tenant_authority_v1(
    uuid,uuid,bigint,bigint,timestamptz
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.analytics_register_tenant_authority_v1(
    uuid,uuid,bigint,bigint,timestamptz
) FROM forgecrud_runtime;
REVOKE EXECUTE ON FUNCTION platform.analytics_register_tenant_authority_v1(
    uuid,uuid,bigint,bigint,timestamptz
) FROM forgecrud_v2_runtime;

WITH retired(signature) AS (
    VALUES
        ('platform.analytics_reconcile_current_tenant_authorities_v1()'),
        ('platform.analytics_register_tenant_authority_v1(uuid,uuid,bigint,bigint,timestamptz)')
), retained(signature) AS (
    VALUES ('platform.analytics_reconcile_current_tenant_authorities_v2()')
), roles(role_name) AS (
    VALUES ('forgecrud_runtime'), ('forgecrud_v2_runtime')
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
              pg_catalog.to_regrole('forgecrud_owner')::oid
           OR NOT routine.prosecdef
           OR NOT COALESCE(routine.proconfig, ARRAY[]::text[])
              @> ARRAY['search_path=pg_catalog']::text[]
           OR pg_catalog.has_function_privilege(
               'forgecrud_runtime', routine.oid, 'EXECUTE'
           )
           OR pg_catalog.has_function_privilege(
               'forgecrud_v2_runtime', routine.oid, 'EXECUTE'
           )
    )
THEN 0 ELSE 1 END;
