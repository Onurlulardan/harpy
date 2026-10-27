-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_reconcile_tenant_schema_runtime_usage_v1()
RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_schema text;
    v_namespace oid;
    v_owner text;
    v_required boolean;
    v_applied bigint := 0;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_migrator'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='tenant schema runtime ACL authority is denied';
    END IF;
    FOR v_schema,v_required IN
        SELECT target.schema_name,pg_catalog.bool_or(target.required)
        FROM (
            SELECT tenant.schema_name::text AS schema_name,false AS required
            FROM platform.tenants AS tenant
            UNION
            SELECT operation.schema_name::text AS schema_name,true AS required
            FROM platform.tenancy_onboarding_operations AS operation
            JOIN platform.tenancy_tenant_migration_jobs AS job
              ON job.operation_id=operation.operation_id
            WHERE job.state='succeeded'
              AND EXISTS (
                  SELECT 1
                  FROM platform.tenancy_tenant_migration_receipts AS receipt
                  WHERE receipt.job_id=job.job_id
                    AND receipt.receipt_kind='migration_completed'
              )
        ) AS target
        GROUP BY target.schema_name ORDER BY target.schema_name
    LOOP
        IF v_schema !~ '^[a-z][a-z0-9_]*$'
           OR pg_catalog.octet_length(v_schema)>63 THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='tenant schema runtime ACL target is invalid';
        END IF;
        SELECT namespace.oid,namespace.nspowner::regrole::text
        INTO v_namespace,v_owner
        FROM pg_catalog.pg_namespace AS namespace
        WHERE namespace.nspname=v_schema;
        IF NOT FOUND THEN
            IF v_required THEN
                RAISE EXCEPTION USING ERRCODE='55000',
                    MESSAGE='tenant schema runtime ACL namespace is missing';
            END IF;
            CONTINUE;
        END IF;
        IF v_owner IS DISTINCT FROM 'harpy_owner' THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='tenant schema runtime ACL ownership is invalid';
        END IF;
        EXECUTE pg_catalog.format(
            'REVOKE CREATE ON SCHEMA %I FROM PUBLIC, harpy_runtime, harpy_v2_runtime',
            v_schema
        );
        EXECUTE pg_catalog.format(
            'GRANT USAGE ON SCHEMA %I TO harpy_runtime',v_schema
        );
        IF NOT pg_catalog.has_schema_privilege(
                'harpy_runtime',v_namespace,'USAGE')
           OR pg_catalog.has_schema_privilege(
                'harpy_runtime',v_namespace,'CREATE')
           OR pg_catalog.has_schema_privilege(
                'harpy_v2_runtime',v_namespace,'CREATE')
           OR EXISTS (
                SELECT 1
                FROM pg_catalog.pg_namespace AS namespace
                CROSS JOIN LATERAL pg_catalog.aclexplode(namespace.nspacl) AS acl
                WHERE namespace.oid=v_namespace AND acl.grantee=0
                  AND acl.privilege_type='CREATE'
           ) THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='tenant schema runtime ACL postcondition is invalid';
        END IF;
        v_applied:=v_applied+1;
    END LOOP;
    RETURN v_applied;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
platform.tenancy_reconcile_tenant_schema_runtime_usage_v1() FROM PUBLIC;
SELECT platform.tenancy_reconcile_tenant_schema_runtime_usage_v1();
DROP FUNCTION platform.tenancy_reconcile_tenant_schema_runtime_usage_v1();
