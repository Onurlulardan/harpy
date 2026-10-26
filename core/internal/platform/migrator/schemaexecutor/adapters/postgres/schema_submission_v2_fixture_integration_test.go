//go:build integration

package postgres

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"testing"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/jackc/pgx/v5/pgxpool"
)

type schemaSubmissionV2Input struct {
	jobID    string
	plan     schemav2.Plan
	prepared schemav2.PreparedPlan
}

type schemaSubmissionV2Receipt struct {
	jobID, digest, state string
	generation           int64
	replay, terminal     bool
}

func prepareSchemaSubmissionV2Input(
	t *testing.T, jobID string, plan schemav2.Plan,
) schemaSubmissionV2Input {
	t.Helper()
	prepared, err := schemav2.PreparePlan(plan)
	if err != nil {
		t.Fatalf("prepare V2 schema submission: %v", err)
	}
	validated, err := schemaexecutor.BuildPlanV2(plan)
	if err != nil || validated.SHA256() != prepared.SHA256() ||
		!bytes.Equal(validated.CanonicalJSON(), prepared.CanonicalJSON()) {
		t.Fatalf("V2 submission/executor preparation parity: %v", err)
	}
	return schemaSubmissionV2Input{jobID: jobID, plan: plan, prepared: prepared}
}

func (input schemaSubmissionV2Input) stepHex() []string {
	steps := input.prepared.Steps()
	encoded := make([]string, len(steps))
	for index := range steps {
		encoded[index] = hex.EncodeToString(steps[index].CanonicalPreimage())
	}
	return encoded
}

func callSchemaSubmissionV2(ctx context.Context, pool *pgxpool.Pool,
	input schemaSubmissionV2Input, planBytes []byte, digest string,
	steps []string) (schemaSubmissionV2Receipt, error) {
	encodedSteps, err := json.Marshal(steps)
	if err != nil {
		return schemaSubmissionV2Receipt{}, err
	}
	var receipt schemaSubmissionV2Receipt
	err = pool.QueryRow(ctx, `SELECT job_id::text,generation,plan_sha256,
replay,job_state,terminal FROM forms_submit_schema_plan_v2(
$1,$2,$3,$4,$5,$6,$7,$8,$9,$10::jsonb)`, input.jobID,
		string(input.plan.Resource.Kind), input.plan.Resource.ID,
		input.plan.Resource.Revision, input.plan.ContractVersion,
		input.plan.Compiler.Version, input.plan.Compiler.PostgreSQLMajor,
		planBytes, digest, string(encodedSteps)).Scan(
		&receipt.jobID, &receipt.generation, &receipt.digest,
		&receipt.replay, &receipt.state, &receipt.terminal,
	)
	return receipt, err
}

func submitSchemaInputV2(ctx context.Context, pool *pgxpool.Pool,
	input schemaSubmissionV2Input) (schemaSubmissionV2Receipt, error) {
	return callSchemaSubmissionV2(ctx, pool, input, input.prepared.CanonicalJSON(),
		input.prepared.SHA256(), input.stepHex())
}

func installSchemaSubmissionV2Wrapper(t *testing.T, ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment, schemaName string) {
	t.Helper()
	transaction, err := environment.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()
	if _, err = transaction.Exec(ctx,
		"SET LOCAL search_path TO "+pgIdentifier(schemaName)+",pg_catalog"); err != nil {
		t.Fatal(err)
	}
	_, err = transaction.Exec(ctx, `
CREATE FUNCTION forms_submit_schema_plan_v2(
    p_job_id uuid,p_resource_kind text,p_resource_id uuid,
    p_resource_revision bigint,p_plan_contract_version text,
    p_compiler_version text,p_postgresql_major integer,
    p_canonical_plan_bytes bytea,p_plan_sha256 text,p_canonical_steps jsonb
)
RETURNS TABLE(job_id uuid,generation bigint,plan_sha256 text,
    replay boolean,job_state text,terminal boolean)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_scope_schema name; v_tenant_id uuid;
BEGIN
    v_scope_schema:=pg_catalog.current_schema()::name;
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='V2 Forms role is invalid';
    END IF;
    SELECT identity_row.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity identity_row
    WHERE identity_row.singleton
      AND identity_row.schema_name=v_scope_schema::text;
    RETURN QUERY SELECT submitted.*
    FROM platform.platform_migrator_submit_schema_plan_v2(
        v_scope_schema,v_tenant_id,p_job_id,p_resource_kind,p_resource_id,
        p_resource_revision,p_plan_contract_version,p_compiler_version,
        p_postgresql_major,p_canonical_plan_bytes,p_plan_sha256,p_canonical_steps
    ) submitted;
EXCEPTION WHEN NO_DATA_FOUND OR TOO_MANY_ROWS THEN
    RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='V2 TenantScope is invalid';
END;
$function$;
REVOKE ALL ON FUNCTION forms_submit_schema_plan_v2(
    uuid,text,uuid,bigint,text,text,integer,bytea,text,jsonb) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION forms_submit_schema_plan_v2(
    uuid,text,uuid,bigint,text,text,integer,bytea,text,jsonb)
TO harpy_v2_runtime;`)
	if err != nil {
		t.Fatalf("install V2 submission wrapper: %v", err)
	}
	if _, err = transaction.Exec(ctx, "GRANT USAGE ON SCHEMA "+
		pgIdentifier(schemaName)+" TO harpy_v2_runtime"); err != nil {
		t.Fatal(err)
	}
	if err = transaction.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}
