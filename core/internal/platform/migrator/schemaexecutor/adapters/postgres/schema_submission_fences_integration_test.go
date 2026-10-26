//go:build integration

package postgres

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func (suite schemaSubmissionSuite) assertReplayAndPendingFences(
	t *testing.T,
	ctx context.Context,
) {
	t.Helper()
	replayInput := prepareSchemaSubmissionInput(t, schemaSubmissionUUID(710), suite.input.plan)
	receipt, err := submitSchemaInput(ctx, suite.runtime, replayInput)
	if err != nil || !receipt.replay || receipt.jobID != suite.input.jobID ||
		receipt.digest != suite.input.prepared.SHA256() {
		t.Fatalf("exact replay receipt = %+v err=%v", receipt, err)
	}
	jobConflict := prepareSchemaSubmissionInput(t, suite.input.jobID,
		schemaExecutorCreatePlan(t, suite.tenantA, schemaSubmissionUUID(711), 1))
	_, err = submitSchemaInput(ctx, suite.runtime, jobConflict)
	requireSchemaSubmissionState(t, err, "55000")
	activeConflict := prepareSchemaSubmissionInput(t, schemaSubmissionUUID(712),
		schemaExecutorCreatePlan(t, suite.tenantA, suite.resourceID, 2))
	_, err = submitSchemaInput(ctx, suite.runtime, activeConflict)
	requireSchemaSubmissionState(t, err, "23505")
}

func (suite schemaSubmissionSuite) assertTamperRollbackAndAuthorityFences(
	t *testing.T,
	ctx context.Context,
) {
	t.Helper()
	hashInput := prepareSchemaSubmissionInput(t, schemaSubmissionUUID(714),
		schemaExecutorCreatePlan(t, suite.tenantA, schemaSubmissionUUID(713), 1))
	wrongDigest := "0" + hashInput.prepared.SHA256()[1:]
	if wrongDigest == hashInput.prepared.SHA256() {
		wrongDigest = "1" + wrongDigest[1:]
	}
	_, err := callSchemaSubmission(ctx, suite.runtime, hashInput,
		hashInput.prepared.CanonicalJSON(), wrongDigest, hashInput.stepHex())
	requireSchemaSubmissionState(t, err, "22023")
	planInput := prepareSchemaSubmissionInput(t, schemaSubmissionUUID(715),
		schemaExecutorCreatePlan(t, suite.tenantA, schemaSubmissionUUID(716), 1))
	tamperedPlan := append(planInput.prepared.CanonicalJSON(), '\n')
	_, err = callSchemaSubmission(ctx, suite.runtime, planInput, tamperedPlan,
		planInput.prepared.SHA256(), planInput.stepHex())
	requireSchemaSubmissionState(t, err, "22023")
	suite.assertAtomicStepRollback(t, ctx)
	suite.assertAuthorityAndScopeFences(t, ctx)
}

func (suite schemaSubmissionSuite) assertAtomicStepRollback(
	t *testing.T,
	ctx context.Context,
) {
	t.Helper()
	input := prepareSchemaSubmissionInput(t, schemaSubmissionUUID(720),
		schemaExecutorCreateAndFieldPlan(t, suite.tenantA,
			schemaSubmissionUUID(718), schemaSubmissionUUID(719)))
	steps := input.stepHex()
	document, err := hex.DecodeString(steps[1])
	if err != nil {
		t.Fatal(err)
	}
	document = bytes.Replace(document, []byte(input.prepared.SHA256()),
		[]byte(strings.Repeat("0", 64)), 1)
	steps[1] = hex.EncodeToString(document)
	_, err = callSchemaSubmission(ctx, suite.runtime, input,
		input.prepared.CanonicalJSON(), input.prepared.SHA256(), steps)
	requireSchemaSubmissionState(t, err, "22023")
	var jobs, stepsStored int
	err = suite.environment.pool.QueryRow(ctx, `SELECT
(SELECT count(*) FROM platform.schema_jobs WHERE job_id=$1),
(SELECT count(*) FROM platform.schema_job_steps WHERE job_id=$1)`, input.jobID).
		Scan(&jobs, &stepsStored)
	if err != nil || jobs != 0 || stepsStored != 0 {
		t.Fatalf("partial submission survived: jobs=%d steps=%d err=%v",
			jobs, stepsStored, err)
	}
}

func (suite schemaSubmissionSuite) assertAuthorityAndScopeFences(
	t *testing.T,
	ctx context.Context,
) {
	t.Helper()
	_, err := suite.environment.pool.Exec(ctx, `UPDATE
platform.schema_submission_authority_fixture SET is_current=FALSE WHERE tenant_id=$1`,
		suite.tenantA)
	if err != nil {
		t.Fatal(err)
	}
	stale := prepareSchemaSubmissionInput(t, schemaSubmissionUUID(722),
		schemaExecutorCreatePlan(t, suite.tenantA, schemaSubmissionUUID(721), 1))
	_, submitErr := submitSchemaInput(ctx, suite.runtime, stale)
	requireSchemaSubmissionState(t, submitErr, "42501")
	_, err = suite.environment.pool.Exec(ctx, `UPDATE
platform.schema_submission_authority_fixture SET is_current=TRUE WHERE tenant_id=$1`,
		suite.tenantA)
	if err != nil {
		t.Fatal(err)
	}
	crossTenant := prepareSchemaSubmissionInput(t, schemaSubmissionUUID(724),
		schemaExecutorCreatePlan(t, suite.tenantB, schemaSubmissionUUID(723), 1))
	_, err = submitSchemaInput(ctx, suite.runtime, crossTenant)
	requireSchemaSubmissionState(t, err, "22023")
	if _, err = suite.environment.pool.Exec(ctx,
		`UPDATE platform.tenants SET status='SUSPENDED' WHERE id=$1`, suite.tenantA); err != nil {
		t.Fatal(err)
	}
	inactive := prepareSchemaSubmissionInput(t, schemaSubmissionUUID(726),
		schemaExecutorCreatePlan(t, suite.tenantA, schemaSubmissionUUID(725), 1))
	_, submitErr = submitSchemaInput(ctx, suite.runtime, inactive)
	requireSchemaSubmissionState(t, submitErr, "42501")
	if _, err = suite.environment.pool.Exec(ctx,
		`UPDATE platform.tenants SET status='ACTIVE' WHERE id=$1`, suite.tenantA); err != nil {
		t.Fatal(err)
	}
}

func (suite schemaSubmissionSuite) assertRevisionMonotonicity(
	t *testing.T,
	ctx context.Context,
) {
	t.Helper()
	fieldID := schemaSubmissionUUID(727)
	stale := prepareSchemaSubmissionInput(t, schemaSubmissionUUID(728),
		schemaExecutorCreateAndFieldPlan(t, suite.tenantA, suite.resourceID, fieldID))
	_, err := submitSchemaInput(ctx, suite.runtime, stale)
	requireSchemaSubmissionState(t, err, "55000")
	higherPlan := schemaExecutorCreateAndFieldPlan(t, suite.tenantA, suite.resourceID, fieldID)
	higherPlan.Resource.Revision = 2
	higher := prepareSchemaSubmissionInput(t, schemaSubmissionUUID(729), higherPlan)
	receipt, err := submitSchemaInput(ctx, suite.runtime, higher)
	if err != nil || receipt.replay || receipt.jobID != higher.jobID {
		t.Fatalf("monotonic revision receipt = %+v err=%v", receipt, err)
	}
}

func requireSchemaSubmissionState(t *testing.T, err error, expected string) {
	t.Helper()
	var pgError *pgconn.PgError
	if !errors.As(err, &pgError) || pgError.Code != expected {
		t.Fatalf("SQL error = %v, want state %s", err, expected)
	}
}

func schemaSubmissionStepsJSON(t *testing.T, input schemaSubmissionInput) string {
	t.Helper()
	encoded, err := json.Marshal(input.stepHex())
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}
