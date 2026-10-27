//go:build integration

package postgres

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/jackc/pgx/v5/pgxpool"
)

type schemaResultReceipt struct {
	jobID         string
	generation    int64
	succeeded     bool
	receiptSHA256 []byte
	failureCode   *string
	failureSHA256 []byte
}

func callSchemaResult(
	ctx context.Context,
	pool *pgxpool.Pool,
	jobID string,
	planSHA256 string,
	minimumGeneration int64,
) (schemaResultReceipt, error) {
	var result schemaResultReceipt
	err := pool.QueryRow(ctx, `SELECT job_id::text, result_generation, succeeded,
receipt_sha256, failure_code, failure_sha256
FROM forms_resolve_schema_result_v1($1,$2,$3)`,
		jobID, planSHA256, minimumGeneration,
	).Scan(&result.jobID, &result.generation, &result.succeeded,
		&result.receiptSHA256, &result.failureCode, &result.failureSHA256)
	return result, err
}

func (suite schemaSubmissionSuite) assertNonterminalResultDenied(
	t *testing.T,
	ctx context.Context,
) {
	t.Helper()
	_, err := callSchemaResult(ctx, suite.runtime, suite.input.jobID,
		suite.input.prepared.SHA256(), 1)
	requireSchemaSubmissionState(t, err, "55000")
}

func (suite schemaSubmissionSuite) assertTerminalResults(
	t *testing.T,
	ctx context.Context,
) {
	t.Helper()
	success, err := callSchemaResult(ctx, suite.runtime, suite.input.jobID,
		suite.input.prepared.SHA256(), 1)
	if err != nil || !success.succeeded || success.jobID != suite.input.jobID ||
		success.generation < 1 || len(success.receiptSHA256) != 32 ||
		success.failureCode != nil || success.failureSHA256 != nil {
		t.Fatalf("successful schema result = %+v err=%v", success, err)
	}
	replayed, err := callSchemaResult(ctx, suite.runtime, suite.input.jobID,
		suite.input.prepared.SHA256(), success.generation)
	if err != nil || !bytes.Equal(replayed.receiptSHA256, success.receiptSHA256) {
		t.Fatalf("schema result receipt is not deterministic: %v", err)
	}
	_, err = callSchemaResult(ctx, suite.runtime, suite.input.jobID,
		suite.input.prepared.SHA256(), success.generation+1)
	requireSchemaSubmissionState(t, err, "55000")
	wrongDigest := "0" + suite.input.prepared.SHA256()[1:]
	if wrongDigest == suite.input.prepared.SHA256() {
		wrongDigest = "1" + wrongDigest[1:]
	}
	_, err = callSchemaResult(ctx, suite.runtime, suite.input.jobID, wrongDigest, 1)
	requireSchemaSubmissionState(t, err, "42501")
	suite.assertFailedAndCrossTenantResults(t, ctx)
}

func (suite schemaSubmissionSuite) assertFailedAndCrossTenantResults(
	t *testing.T,
	ctx context.Context,
) {
	t.Helper()
	failedInput := prepareSchemaSubmissionInput(t, schemaSubmissionUUID(731),
		schemaExecutorCreatePlan(t, suite.tenantA, schemaSubmissionUUID(730), 1))
	if _, err := submitSchemaInput(ctx, suite.runtime, failedInput); err != nil {
		t.Fatal(err)
	}
	record, found, err := suite.environment.store.Claim(ctx, schemaexecutor.ClaimRequest{
		ExecutorID: schemaSubmissionUUID(732), LeaseID: schemaSubmissionUUID(733),
	})
	if err != nil || !found {
		t.Fatalf("claim failure fixture: found=%t err=%v", found, err)
	}
	lease, _, err := schemaexecutor.AcceptIntegrationClaim(record)
	if err != nil {
		t.Fatal(err)
	}
	if err = suite.environment.store.MarkRunning(ctx, lease); err != nil {
		t.Fatal(err)
	}
	err = suite.environment.store.RecordFailure(ctx, schemaexecutor.FailureRequest{
		Lease: lease, State: schemaexecutor.FailureTerminal, Code: "ddl_rejected",
		RedactedDetailSHA256: strings.Repeat("a", 64),
	})
	if err != nil {
		t.Fatal(err)
	}
	failed, err := callSchemaResult(ctx, suite.runtime, failedInput.jobID,
		failedInput.prepared.SHA256(), 1)
	if err != nil || failed.succeeded || failed.failureCode == nil ||
		*failed.failureCode != "SCHEMA_JOB_FAILED" ||
		len(failed.receiptSHA256) != 32 ||
		!bytes.Equal(failed.failureSHA256, bytes.Repeat([]byte{0xaa}, 32)) {
		t.Fatalf("failed schema result = %+v err=%v", failed, err)
	}
	again, err := callSchemaResult(ctx, suite.runtime, failedInput.jobID,
		failedInput.prepared.SHA256(), failed.generation)
	if err != nil || !bytes.Equal(again.receiptSHA256, failed.receiptSHA256) ||
		!bytes.Equal(again.failureSHA256, failed.failureSHA256) {
		t.Fatalf("failed schema result is not deterministic: %v", err)
	}
	suite.assertCancelledResult(t, ctx)
	runtimeB := openSchemaSubmissionRuntime(t, ctx, suite.environment, "submission_b")
	inputB := prepareSchemaSubmissionInput(t, schemaSubmissionUUID(735),
		schemaExecutorCreatePlan(t, suite.tenantB, schemaSubmissionUUID(734), 1))
	if _, err = submitSchemaInput(ctx, runtimeB, inputB); err != nil {
		t.Fatal(err)
	}
	_, err = callSchemaResult(ctx, suite.runtime, inputB.jobID,
		inputB.prepared.SHA256(), 1)
	requireSchemaSubmissionState(t, err, "42501")
}

func (suite schemaSubmissionSuite) assertCancelledResult(
	t *testing.T,
	ctx context.Context,
) {
	t.Helper()
	input := prepareSchemaSubmissionInput(t, schemaSubmissionUUID(737),
		schemaExecutorCreatePlan(t, suite.tenantA, schemaSubmissionUUID(736), 1))
	if _, err := submitSchemaInput(ctx, suite.runtime, input); err != nil {
		t.Fatal(err)
	}
	_, err := suite.environment.pool.Exec(ctx, `WITH authoritative_time AS (
SELECT pg_catalog.clock_timestamp() AS now
)
UPDATE platform.schema_jobs AS job SET state='cancelled',
generation=job.generation+1, next_executor='none', updated_at=authoritative_time.now,
finished_at=authoritative_time.now, failure_code='cancelled_by_owner'
FROM authoritative_time WHERE job.job_id=$1 AND job.state='queued'`, input.jobID)
	if err != nil {
		t.Fatal(err)
	}
	result, err := callSchemaResult(ctx, suite.runtime, input.jobID,
		input.prepared.SHA256(), 1)
	if err != nil || result.succeeded || result.failureCode == nil ||
		*result.failureCode != "SCHEMA_JOB_CANCELLED" ||
		len(result.receiptSHA256) != 32 || len(result.failureSHA256) != 32 {
		t.Fatalf("cancelled schema result = %+v err=%v", result, err)
	}
	replayed, err := callSchemaResult(ctx, suite.runtime, input.jobID,
		input.prepared.SHA256(), result.generation)
	if err != nil || !bytes.Equal(replayed.receiptSHA256, result.receiptSHA256) ||
		!bytes.Equal(replayed.failureSHA256, result.failureSHA256) {
		t.Fatalf("cancelled schema result is not deterministic: %v", err)
	}
}
