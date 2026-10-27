//go:build integration

package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/jackc/pgx/v5/pgxpool"
)

type schemaSubmissionV2Suite struct {
	environment            *schemaExecutorIntegrationEnvironment
	runtime                *pgxpool.Pool
	tenantA, tenantB       string
	authorityA, authorityB string
	input                  schemaSubmissionV2Input
}

func TestSchemaPlanSubmissionV2AuthorityOnPostgreSQL184(t *testing.T) {
	for _, variant := range []string{"fresh", "v1_adopted"} {
		t.Run(variant, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
			defer cancel()
			runSchemaSubmissionV2Suite(t, ctx, variant)
		})
	}
}

func runSchemaSubmissionV2Suite(t *testing.T, ctx context.Context, variant string) {
	t.Helper()
	environment := startSchemaExecutorIntegrationVariant(t, ctx, variant)
	tenantA, tenantB := v2ExactTenant, schemaSubmissionUUID(891)
	authorityA, authorityB := schemaSubmissionUUID(892), schemaSubmissionUUID(893)
	environment.createTenant(t, ctx, tenantA, "submission_v2")
	environment.createTenant(t, ctx, tenantB, "submission_v2_b")
	installSchemaSubmissionV2Authority(t, ctx, environment, tenantA, authorityA)
	installSchemaSubmissionV2Wrapper(t, ctx, environment, "submission_v2")
	suite := schemaSubmissionV2Suite{environment: environment,
		runtime: openSchemaSubmissionRuntime(t, ctx, environment, "submission_v2"),
		tenantA: tenantA, tenantB: tenantB, authorityA: authorityA,
		authorityB: authorityB,
		input:      prepareSchemaSubmissionV2Input(t, schemaSubmissionUUID(894), v2ExactPlan(t))}
	suite.assertInitialSnapshot(t, ctx)
	suite.assertReplayAndTransplantFences(t, ctx)
	suite.assertAuthorityRotationFence(t, ctx)
	suite.assertPrivateHelperDenied(t, ctx)
}

func (suite schemaSubmissionV2Suite) assertInitialSnapshot(
	t *testing.T, ctx context.Context,
) {
	t.Helper()
	receipt, err := submitSchemaInputV2(ctx, suite.runtime, suite.input)
	if err != nil || receipt.jobID != suite.input.jobID || receipt.generation != 1 ||
		receipt.digest != suite.input.prepared.SHA256() || receipt.replay ||
		receipt.state != "queued" || receipt.terminal {
		t.Fatalf("initial V2 submission = %+v err=%v", receipt, err)
	}
	var authorityID, groupID, contract, compiler string
	var groupGeneration, billingGeneration, formsGeneration int64
	var canonical []byte
	err = suite.environment.pool.QueryRow(ctx, `SELECT
submission_authority_id::text,submission_authority_group_id,
submission_group_generation,submission_billing_generation,
submission_forms_generation,plan_contract_version,compiler_version,
canonical_plan_bytes FROM platform.schema_jobs WHERE job_id=$1`,
		suite.input.jobID).Scan(&authorityID, &groupID, &groupGeneration,
		&billingGeneration, &formsGeneration, &contract, &compiler, &canonical)
	if err != nil || authorityID != suite.authorityA || groupID != "billing_forms" ||
		groupGeneration != schemaSubmissionV2GroupGeneration ||
		billingGeneration != schemaSubmissionV2BillingGeneration ||
		formsGeneration != schemaSubmissionV2FormsGeneration ||
		contract != "forms-schema-plan.v2" || compiler != "2.0.0" ||
		!bytes.Equal(canonical, suite.input.prepared.CanonicalJSON()) {
		t.Fatalf("frozen V2 submission evidence differs: %v", err)
	}
	_, err = suite.environment.pool.Exec(ctx, `UPDATE platform.schema_jobs
SET submission_authority_id=$2 WHERE job_id=$1`, suite.input.jobID, suite.authorityB)
	requireSchemaSubmissionState(t, err, "55000")
}

func (suite schemaSubmissionV2Suite) assertReplayAndTransplantFences(
	t *testing.T, ctx context.Context,
) {
	t.Helper()
	replay := suite.input
	replay.jobID = schemaSubmissionUUID(895)
	receipt, err := submitSchemaInputV2(ctx, suite.runtime, replay)
	if err != nil || !receipt.replay || receipt.jobID != suite.input.jobID ||
		receipt.state != "queued" || receipt.terminal {
		t.Fatalf("exact V2 replay = %+v err=%v", receipt, err)
	}
	transplantPlan := suite.input.plan
	transplantPlan.Resource.Revision++
	transplant := prepareSchemaSubmissionV2Input(t, schemaSubmissionUUID(896), transplantPlan)
	request := suite.input
	request.jobID = transplant.jobID
	_, err = callSchemaSubmissionV2(ctx, suite.runtime, request,
		suite.input.prepared.CanonicalJSON(), suite.input.prepared.SHA256(),
		transplant.stepHex())
	requireSchemaSubmissionState(t, err, "22023")
	crossPlan := suite.input.plan
	crossPlan.TenantID = suite.tenantB
	cross := prepareSchemaSubmissionV2Input(t, schemaSubmissionUUID(897), crossPlan)
	_, err = submitSchemaInputV2(ctx, suite.runtime, cross)
	requireSchemaSubmissionState(t, err, "22023")
	var inserted int
	if err = suite.environment.pool.QueryRow(ctx, `SELECT count(*)
FROM platform.schema_jobs WHERE job_id IN ($1,$2)`, transplant.jobID, cross.jobID).
		Scan(&inserted); err != nil || inserted != 0 {
		t.Fatalf("rejected V2 transplant survived: count=%d err=%v", inserted, err)
	}
}

func (suite schemaSubmissionV2Suite) assertAuthorityRotationFence(
	t *testing.T, ctx context.Context,
) {
	t.Helper()
	if !suite.authorityCurrent(t, ctx) {
		t.Fatal("frozen V2 authority was not initially current")
	}
	rotateSchemaSubmissionV2Authority(t, ctx, suite.environment,
		suite.tenantA, suite.authorityB)
	if suite.authorityCurrent(t, ctx) {
		t.Fatal("transplanted V2 authority was accepted as the frozen authority")
	}
	replay := suite.input
	replay.jobID = schemaSubmissionUUID(898)
	_, err := submitSchemaInputV2(ctx, suite.runtime, replay)
	requireSchemaSubmissionState(t, err, "55000")
	_, found, err := suite.environment.store.Claim(ctx, schemaexecutor.ClaimRequest{
		ExecutorID: schemaSubmissionUUID(899), LeaseID: schemaSubmissionUUID(900),
	})
	if err != nil || found {
		t.Fatalf("stale V2 authority job claim: found=%t err=%v", found, err)
	}
}

func (suite schemaSubmissionV2Suite) authorityCurrent(t *testing.T, ctx context.Context) bool {
	t.Helper()
	var current bool
	err := suite.environment.pool.QueryRow(ctx, `SELECT
platform.platform_migrator_schema_authority_current_v2($1,$2,$3,$4,$5,$6)`,
		suite.tenantA, suite.authorityA, "billing_forms",
		schemaSubmissionV2GroupGeneration, schemaSubmissionV2BillingGeneration,
		schemaSubmissionV2FormsGeneration).Scan(&current)
	if err != nil {
		t.Fatalf("read V2 authority fence: %v", err)
	}
	return current
}

func (suite schemaSubmissionV2Suite) assertPrivateHelperDenied(
	t *testing.T, ctx context.Context,
) {
	t.Helper()
	var executable bool
	err := suite.environment.pool.QueryRow(ctx, `SELECT pg_catalog.has_function_privilege(
'harpy_v2_runtime','platform.platform_migrator_submit_schema_plan_v2(name,uuid,uuid,text,uuid,bigint,text,text,integer,bytea,text,jsonb)','EXECUTE')`).
		Scan(&executable)
	if err != nil || executable {
		t.Fatalf("V2 private helper ACL: executable=%t err=%v", executable, err)
	}
	steps, err := json.Marshal(suite.input.stepHex())
	if err != nil {
		t.Fatal(err)
	}
	_, err = suite.runtime.Exec(ctx, `SELECT * FROM
platform.platform_migrator_submit_schema_plan_v2(
$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12::jsonb)`, "submission_v2",
		suite.tenantA, suite.input.jobID, string(suite.input.plan.Resource.Kind),
		suite.input.plan.Resource.ID, suite.input.plan.Resource.Revision,
		suite.input.plan.ContractVersion, suite.input.plan.Compiler.Version,
		suite.input.plan.Compiler.PostgreSQLMajor, suite.input.prepared.CanonicalJSON(),
		suite.input.prepared.SHA256(), string(steps))
	requireSchemaSubmissionState(t, err, "42501")
}
