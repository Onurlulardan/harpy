//go:build integration

package postgres

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/jackc/pgx/v5/pgxpool"
)

type schemaSubmissionSuite struct {
	environment *schemaExecutorIntegrationEnvironment
	runtime     *pgxpool.Pool
	tenantA     string
	tenantB     string
	resourceID  string
	input       schemaSubmissionInput
}

func TestSchemaPlanSubmissionOnPostgreSQL184(t *testing.T) {
	for _, variant := range []string{"fresh", "v1_adopted"} {
		t.Run(variant, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
			defer cancel()
			runSchemaSubmissionSuite(t, ctx, variant)
		})
	}
}

func runSchemaSubmissionSuite(t *testing.T, ctx context.Context, variant string) {
	t.Helper()
	environment := startSchemaExecutorIntegrationVariant(t, ctx, variant)
	tenantA, tenantB := schemaSubmissionUUID(701), schemaSubmissionUUID(702)
	environment.createTenant(t, ctx, tenantA, "submission_a")
	environment.createTenant(t, ctx, tenantB, "submission_b")
	installSchemaSubmissionAuthorityFixture(t, ctx, environment, tenantA, tenantB)
	installSchemaSubmissionWrapper(t, ctx, environment, "submission_a")
	installSchemaSubmissionWrapper(t, ctx, environment, "submission_b")
	installSchemaResultWrapper(t, ctx, environment, "submission_a")
	installSchemaResultWrapper(t, ctx, environment, "submission_b")
	resourceID := schemaSubmissionUUID(703)
	input := prepareSchemaSubmissionInput(t, schemaSubmissionUUID(704),
		schemaExecutorCreatePlan(t, tenantA, resourceID, 1))
	suite := schemaSubmissionSuite{
		environment: environment,
		runtime:     openSchemaSubmissionRuntime(t, ctx, environment, "submission_a"),
		tenantA:     tenantA, tenantB: tenantB, resourceID: resourceID, input: input,
	}
	suite.assertInitialSubmission(t, ctx)
	suite.assertReplayAndPendingFences(t, ctx)
	suite.assertTamperRollbackAndAuthorityFences(t, ctx)
	suite.assertNonterminalResultDenied(t, ctx)
	suite.assertExecutorRoundTrip(t, ctx)
	suite.assertTerminalResults(t, ctx)
	suite.assertRevisionMonotonicity(t, ctx)
	suite.assertDirectAccessDenied(t, ctx)
}

func (suite schemaSubmissionSuite) assertInitialSubmission(
	t *testing.T,
	ctx context.Context,
) {
	t.Helper()
	receipt, err := submitSchemaInput(ctx, suite.runtime, suite.input)
	if err != nil {
		t.Fatalf("submit canonical schema plan: %v", err)
	}
	if receipt.jobID != suite.input.jobID || receipt.generation != 1 || receipt.replay ||
		receipt.digest != suite.input.prepared.SHA256() {
		t.Fatalf("submission receipt = %+v", receipt)
	}
	var planBytes, stepBytes []byte
	var stepCount int
	err = suite.environment.pool.QueryRow(ctx, `SELECT job.canonical_plan_bytes,
step.canonical_step_bytes, job.step_count FROM platform.schema_jobs AS job
JOIN platform.schema_job_steps AS step ON step.job_id=job.job_id AND step.ordinal=0
WHERE job.job_id=$1`, suite.input.jobID).Scan(&planBytes, &stepBytes, &stepCount)
	if err != nil || stepCount != 1 ||
		!bytes.Equal(planBytes, suite.input.prepared.CanonicalJSON()) ||
		!bytes.Equal(stepBytes, suite.input.prepared.Steps()[0].CanonicalPreimage()) {
		t.Fatalf("persisted canonical evidence differs: count=%d err=%v", stepCount, err)
	}
}

func (suite schemaSubmissionSuite) assertExecutorRoundTrip(
	t *testing.T,
	ctx context.Context,
) {
	t.Helper()
	record, found, err := suite.environment.store.Claim(ctx, schemaexecutor.ClaimRequest{
		ExecutorID: schemaSubmissionUUID(705), LeaseID: schemaSubmissionUUID(706),
	})
	if err != nil || !found {
		t.Fatalf("claim submitted plan: found=%t err=%v", found, err)
	}
	lease, validated, err := schemaexecutor.AcceptIntegrationClaim(record)
	step := suite.input.prepared.Steps()[0]
	if err != nil || validated.StepCount() != 1 ||
		validated.CurrentStep().HashSHA256() != step.SHA256() ||
		!bytes.Equal(record.Job.CanonicalPlan, suite.input.prepared.CanonicalJSON()) {
		t.Fatalf("submit-to-executor validation differs: %v", err)
	}
	if err = suite.environment.store.MarkRunning(ctx, lease); err != nil {
		t.Fatal(err)
	}
	if err = suite.environment.store.Apply(ctx, schemaexecutor.ApplyRequest{
		Lease: lease, Plan: validated,
	}); err != nil {
		t.Fatalf("apply submitted plan: %v", err)
	}
}
