//go:build integration

package postgres

import (
	"context"
	"testing"
)

func installSchemaResultWrapper(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	schemaName string,
) {
	t.Helper()
	transaction, err := environment.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()
	if _, err = transaction.Exec(ctx,
		"SET LOCAL search_path TO "+pgIdentifier(schemaName)+", pg_catalog"); err != nil {
		t.Fatal(err)
	}
	_, err = transaction.Exec(ctx, `
CREATE FUNCTION forms_resolve_schema_result_v1(
    p_job_id uuid,
    p_plan_sha256 text,
    p_minimum_generation bigint
)
RETURNS TABLE(
    job_id uuid,
    result_generation bigint,
    succeeded boolean,
    receipt_sha256 bytea,
    failure_code text,
    failure_sha256 bytea
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_schema name;
    v_tenant_id uuid;
BEGIN
    v_schema := pg_catalog.current_schema()::name;
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Forms result role is invalid';
    END IF;
    BEGIN
        SELECT identity_row.tenant_id INTO STRICT v_tenant_id
        FROM tenancy_schema_identity AS identity_row
        WHERE identity_row.singleton
          AND identity_row.schema_name = v_schema::text;
    EXCEPTION WHEN NO_DATA_FOUND OR TOO_MANY_ROWS THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Forms result TenantScope is invalid';
    END;
    RETURN QUERY SELECT resolved.job_id, resolved.result_generation,
        resolved.succeeded, resolved.receipt_sha256, resolved.failure_code,
        resolved.failure_sha256
    FROM platform.platform_migrator_resolve_schema_result_v1(
        v_tenant_id, p_job_id, p_plan_sha256, p_minimum_generation
    ) AS resolved;
END;
$function$;
REVOKE ALL ON FUNCTION
forms_resolve_schema_result_v1(uuid, text, bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION
forms_resolve_schema_result_v1(uuid, text, bigint) TO harpy_v2_runtime;`)
	if err != nil {
		t.Fatalf("install Forms result wrapper: %v", err)
	}
	if err = transaction.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}
