//go:build integration

package postgres

import (
	"context"
	"testing"
)

func installSchemaSubmissionAuthorityFixture(
	t *testing.T,
	ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment,
	tenantIDs ...string,
) {
	t.Helper()
	_, err := environment.pool.Exec(ctx, `
CREATE TABLE platform.schema_submission_authority_fixture (
    tenant_id uuid PRIMARY KEY,
    is_current boolean NOT NULL
);
REVOKE ALL ON platform.schema_submission_authority_fixture FROM PUBLIC;
CREATE FUNCTION platform.billing_resolve_current_tenant_authority_v1(p_tenant_id uuid)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR NOT EXISTS (
            SELECT 1 FROM platform.schema_submission_authority_fixture AS fixture
            WHERE fixture.tenant_id = p_tenant_id AND fixture.is_current
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing+forms tenant authority is not current';
    END IF;
    RETURN p_tenant_id;
END;
$function$;
REVOKE ALL ON FUNCTION
platform.billing_resolve_current_tenant_authority_v1(uuid) FROM PUBLIC;
GRANT USAGE ON SCHEMA platform TO harpy_v2_runtime;`)
	if err != nil {
		t.Fatalf("install submission authority fixture: %v", err)
	}
	for _, tenantID := range tenantIDs {
		if _, err = environment.pool.Exec(ctx, `INSERT INTO
platform.schema_submission_authority_fixture VALUES ($1, TRUE)`, tenantID); err != nil {
			t.Fatalf("seed submission authority fixture: %v", err)
		}
	}
}

func installSchemaSubmissionWrapper(
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
CREATE FUNCTION forms_submit_schema_plan_v1(
    p_job_id uuid,
    p_resource_kind text,
    p_resource_id uuid,
    p_resource_revision bigint,
    p_plan_contract_version text,
    p_compiler_version text,
    p_postgresql_major integer,
    p_canonical_plan_bytes bytea,
    p_plan_sha256 text,
    p_canonical_steps jsonb
)
RETURNS TABLE(job_id uuid, generation bigint, plan_sha256 text, replay boolean)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_scope_schema name;
    v_tenant_id uuid;
BEGIN
    v_scope_schema := pg_catalog.current_schema()::name;
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Forms submission role is invalid';
    END IF;
    BEGIN
        SELECT identity_row.tenant_id INTO STRICT v_tenant_id
        FROM tenancy_schema_identity AS identity_row
        WHERE identity_row.singleton
          AND identity_row.schema_name = v_scope_schema::text;
    EXCEPTION WHEN NO_DATA_FOUND OR TOO_MANY_ROWS THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Forms submission TenantScope is invalid';
    END;
    RETURN QUERY SELECT submitted.job_id, submitted.generation,
        submitted.plan_sha256, submitted.replay
    FROM platform.platform_migrator_submit_schema_plan_v1(
        v_scope_schema, v_tenant_id, p_job_id, p_resource_kind, p_resource_id,
        p_resource_revision, p_plan_contract_version, p_compiler_version,
        p_postgresql_major, p_canonical_plan_bytes, p_plan_sha256,
        p_canonical_steps
    ) AS submitted;
END;
$function$;
REVOKE ALL ON FUNCTION forms_submit_schema_plan_v1(
    uuid, text, uuid, bigint, text, text, integer, bytea, text, jsonb
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION forms_submit_schema_plan_v1(
    uuid, text, uuid, bigint, text, text, integer, bytea, text, jsonb
) TO harpy_v2_runtime;`)
	if err != nil {
		t.Fatalf("install Forms submission wrapper: %v", err)
	}
	if _, err = transaction.Exec(ctx, "GRANT USAGE ON SCHEMA "+
		pgIdentifier(schemaName)+" TO harpy_v2_runtime"); err != nil {
		t.Fatal(err)
	}
	if err = transaction.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}
