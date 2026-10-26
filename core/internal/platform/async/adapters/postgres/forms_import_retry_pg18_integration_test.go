//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type formsImportRetryCall struct {
	tenantID, importID, sourceJobID uuid.UUID
	retryJobID, receiptID, actorID  uuid.UUID
	key, requestSHA256              string
	createdAt                       time.Time
}

type formsImportRetryResult struct {
	jobID, receiptID, retryOf uuid.UUID
	replayed                  bool
}

func TestFormsImportRetryPrimitiveOnPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	database := startAdoptionPG18(t, ctx, "forms_import_retry")
	prepareFormsImportRetryPG18(t, ctx, database)
	tenantID, actorID := uuid.New(), uuid.New()
	importID, sourceJobID := mustRetryV7(t), mustRetryV7(t)
	sourceReceiptID := mustRetryV7(t)
	seedFormsImportRetrySource(t, ctx, database, tenantID, importID,
		sourceJobID, actorID, sourceReceiptID, "dead", false)
	runtime := openAdoptionRole(t, ctx, database,
		"harpy_v2_runtime", "retry_runtime_only")
	call := formsImportRetryCall{
		tenantID: tenantID, importID: importID, sourceJobID: sourceJobID,
		retryJobID: mustRetryV7(t), receiptID: mustRetryV7(t), actorID: actorID,
		key: "resume-1", requestSHA256: strings.Repeat("b", 64),
		createdAt: time.Date(2026, 8, 1, 12, 1, 0, 0, time.UTC),
	}
	result, err := invokeFormsImportRetry(ctx, runtime, call, false)
	if err != nil || result.jobID != call.retryJobID ||
		result.receiptID != call.receiptID || result.retryOf != sourceJobID {
		t.Fatalf("retry result = %#v, %v", result, err)
	}
	assertFormsImportRetryRows(t, ctx, database, call, result)
	replayCall := call
	replayCall.retryJobID, replayCall.receiptID = mustRetryV7(t), mustRetryV7(t)
	replay, err := invokeFormsImportRetry(ctx, runtime, replayCall, false)
	if err != nil || !replay.replayed || replay.jobID != result.jobID ||
		replay.receiptID != result.receiptID || replay.retryOf != sourceJobID {
		t.Fatalf("retry replay = %#v, %v", replay, err)
	}
	wrongSource := replayCall
	wrongSource.sourceJobID = mustRetryV7(t)
	_, err = invokeFormsImportRetry(ctx, runtime, wrongSource, false)
	assertRetrySQLState(t, err, "P6048")
	reused := replayCall
	reused.requestSHA256 = strings.Repeat("c", 64)
	_, err = invokeFormsImportRetry(ctx, runtime, reused, false)
	assertRetrySQLState(t, err, "P6041")
	_, err = invokeFormsImportRetry(ctx, runtime, replayCall, true)
	assertRetrySQLState(t, err, "42501")
	assertFormsImportRetryConcurrency(t, ctx, database, runtime,
		tenantID, actorID)
	assertFormsImportRetryLineage(t, ctx, database, runtime, call, result)
	assertFormsImportRetryDenials(t, ctx, database, runtime, tenantID, actorID)
	assertFormsImportRetryScopeDenials(t, ctx, database, runtime,
		tenantID, actorID)
}

func assertFormsImportRetryRows(
	t *testing.T, ctx context.Context, database *sql.DB,
	call formsImportRetryCall, result formsImportRetryResult,
) {
	t.Helper()
	var state, dedupe, operation string
	var retryOf, principal uuid.UUID
	var attempts, generation int
	err := database.QueryRowContext(ctx, `
SELECT job.state,job.dedupe_key,job.retry_of_job_id,job.attempts,
       job.lease_generation,receipt.principal_id,receipt.operation_id
FROM platform.jobs AS job
JOIN platform.idempotency_receipts AS receipt ON receipt.async_job_id=job.id
WHERE job.id=$1 AND receipt.receipt_id=$2`, result.jobID, result.receiptID).Scan(
		&state, &dedupe, &retryOf, &attempts, &generation, &principal, &operation)
	if err != nil || state != "queued" ||
		dedupe != "forms.record-import.retry:"+call.sourceJobID.String() ||
		retryOf != call.sourceJobID || attempts != 0 || generation != 0 ||
		principal != call.actorID || operation != "forms.records.import.resume" {
		t.Fatalf("persisted retry = %s %s %s %d/%d %s %s, %v",
			state, dedupe, retryOf, attempts, generation, principal, operation, err)
	}
}

func assertFormsImportRetryDenials(
	t *testing.T, ctx context.Context, database *sql.DB,
	runtime *pgxpool.Pool, tenantID, actorID uuid.UUID,
) {
	t.Helper()
	for index, state := range []string{"queued", "running", "succeeded", "cancelled"} {
		importID, jobID := mustRetryV7(t), mustRetryV7(t)
		seedFormsImportRetrySource(t, ctx, database, tenantID, importID,
			jobID, actorID, mustRetryV7(t), state, false)
		call := denialRetryCall(t, tenantID, importID, jobID, actorID, index)
		_, err := invokeFormsImportRetry(ctx, runtime, call, false)
		assertRetrySQLState(t, err, "P6045")
	}
	importID, jobID := mustRetryV7(t), mustRetryV7(t)
	seedFormsImportRetrySource(t, ctx, database, tenantID, importID,
		jobID, actorID, mustRetryV7(t), "dead", true)
	call := denialRetryCall(t, tenantID, importID, jobID, actorID, 10)
	_, err := invokeFormsImportRetry(ctx, runtime, call, false)
	assertRetrySQLState(t, err, "P6045")
	importID, jobID = mustRetryV7(t), mustRetryV7(t)
	seedFormsImportRetrySource(t, ctx, database, tenantID, importID,
		jobID, actorID, mustRetryV7(t), "dead", false)
	call = denialRetryCall(t, tenantID, importID, jobID, uuid.New(), 11)
	_, err = invokeFormsImportRetry(ctx, runtime, call, false)
	assertRetrySQLState(t, err, "42501")
}

func denialRetryCall(
	t *testing.T, tenantID, importID, jobID, actorID uuid.UUID, index int,
) formsImportRetryCall {
	t.Helper()
	return formsImportRetryCall{tenantID: tenantID, importID: importID,
		sourceJobID: jobID, retryJobID: mustRetryV7(t), receiptID: mustRetryV7(t),
		actorID: actorID, key: fmt.Sprintf("deny-%d", index),
		requestSHA256: strings.Repeat("d", 64),
		createdAt:     time.Date(2026, 8, 1, 12, 2, 0, 0, time.UTC)}
}

func invokeFormsImportRetry(
	ctx context.Context, runtime *pgxpool.Pool, call formsImportRetryCall, direct bool,
) (formsImportRetryResult, error) {
	name := "public.retry_forms_import"
	if direct {
		name = "platform.platform_async_forms_import_retry_v1"
	}
	var result formsImportRetryResult
	err := runtime.QueryRow(ctx, "SELECT * FROM "+name+
		"($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)", call.tenantID, call.importID,
		call.sourceJobID, call.retryJobID, call.receiptID, call.actorID, int64(7),
		call.key, call.requestSHA256, call.createdAt).Scan(
		&result.jobID, &result.receiptID, &result.retryOf, &result.replayed)
	return result, err
}

func assertRetrySQLState(t *testing.T, err error, want string) {
	t.Helper()
	var databaseError *pgconn.PgError
	if !errors.As(err, &databaseError) || databaseError.Code != want {
		t.Fatalf("retry error = %v, want SQLSTATE %s", err, want)
	}
}

func mustRetryV7(t *testing.T) uuid.UUID {
	t.Helper()
	value, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return value
}
