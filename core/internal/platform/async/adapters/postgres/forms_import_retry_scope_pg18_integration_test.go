//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func assertFormsImportRetryScopeDenials(
	t *testing.T, ctx context.Context, database *sql.DB, runtime *pgxpool.Pool,
	tenantID, actorID uuid.UUID,
) {
	t.Helper()
	importID, jobID := mustRetryV7(t), mustRetryV7(t)
	seedFormsImportRetrySource(t, ctx, database, tenantID, importID,
		jobID, actorID, mustRetryV7(t), "dead", false)
	call := denialRetryCall(t, tenantID, importID, jobID, actorID, 20)
	call.tenantID = uuid.New()
	_, err := invokeFormsImportRetry(ctx, runtime, call, false)
	assertRetrySQLState(t, err, "P6044")

	importID, jobID = mustRetryV7(t), mustRetryV7(t)
	seedFormsImportRetrySource(t, ctx, database, tenantID, importID,
		jobID, actorID, mustRetryV7(t), "dead", false)
	if _, err = database.ExecContext(ctx,
		"UPDATE platform.jobs SET kind='forms.other' WHERE id=$1", jobID); err != nil {
		t.Fatal(err)
	}
	call = denialRetryCall(t, tenantID, importID, jobID, actorID, 21)
	_, err = invokeFormsImportRetry(ctx, runtime, call, false)
	assertRetrySQLState(t, err, "P6044")

	importID, jobID = mustRetryV7(t), mustRetryV7(t)
	seedFormsImportRetrySource(t, ctx, database, tenantID, importID,
		jobID, actorID, mustRetryV7(t), "dead", false)
	_, err = database.ExecContext(ctx,
		"UPDATE platform.jobs SET authority_group_generation=2 WHERE id=$1", jobID)
	if err == nil {
		_, err = database.ExecContext(ctx, `
UPDATE platform.idempotency_receipts
 SET authority_group_generation=2 WHERE async_job_id=$1`, jobID)
	}
	if err != nil {
		t.Fatal(err)
	}
	call = denialRetryCall(t, tenantID, importID, jobID, actorID, 22)
	_, err = invokeFormsImportRetry(ctx, runtime, call, false)
	assertRetrySQLState(t, err, "42501")

	for index, mutation := range []string{
		"effect_started_at=clock_timestamp()",
		"payload_ref=NULL,payload_redacted_at=clock_timestamp()",
	} {
		importID, jobID = mustRetryV7(t), mustRetryV7(t)
		seedFormsImportRetrySource(t, ctx, database, tenantID, importID,
			jobID, actorID, mustRetryV7(t), "dead", false)
		if _, err = database.ExecContext(ctx,
			"UPDATE platform.jobs SET "+mutation+" WHERE id=$1", jobID); err != nil {
			t.Fatal(err)
		}
		call = denialRetryCall(t, tenantID, importID, jobID, actorID, 23+index)
		_, err = invokeFormsImportRetry(ctx, runtime, call, false)
		assertRetrySQLState(t, err, "P6045")
	}
	assertFormsImportRetryChain(t, ctx, database, runtime, tenantID, actorID)
	assertFormsImportRetryReplayAuthority(t, ctx, database, runtime,
		tenantID, actorID)
}

func assertFormsImportRetryChain(
	t *testing.T, ctx context.Context, database *sql.DB, runtime *pgxpool.Pool,
	tenantID, actorID uuid.UUID,
) {
	t.Helper()
	importID, sourceJobID := mustRetryV7(t), mustRetryV7(t)
	seedFormsImportRetrySource(t, ctx, database, tenantID, importID,
		sourceJobID, actorID, mustRetryV7(t), "dead", false)
	createdAt := time.Date(2026, 8, 1, 12, 4, 0, 0, time.UTC)
	firstCall := formsImportRetryCall{
		tenantID: tenantID, importID: importID, sourceJobID: sourceJobID,
		retryJobID: mustRetryV7(t), receiptID: mustRetryV7(t), actorID: actorID,
		key: "chain-1", requestSHA256: strings.Repeat("e", 64),
		createdAt: createdAt,
	}
	first, err := invokeFormsImportRetry(ctx, runtime, firstCall, false)
	if err != nil {
		t.Fatalf("first chained retry: %v", err)
	}
	if _, err = database.ExecContext(ctx, `
UPDATE platform.jobs SET state='dead',attempts=max_attempts,
 updated_at=$2 WHERE id=$1`, first.jobID, createdAt.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	secondCall := firstCall
	secondCall.sourceJobID = first.jobID
	secondCall.retryJobID, secondCall.receiptID = mustRetryV7(t), mustRetryV7(t)
	secondCall.key, secondCall.requestSHA256 = "chain-2", strings.Repeat("f", 64)
	secondCall.createdAt = createdAt.Add(2 * time.Second)
	second, err := invokeFormsImportRetry(ctx, runtime, secondCall, false)
	if err != nil || second.retryOf != first.jobID || second.jobID != secondCall.retryJobID {
		t.Fatalf("second chained retry = %#v, %v", second, err)
	}
}

func assertFormsImportRetryReplayAuthority(
	t *testing.T, ctx context.Context, database *sql.DB, runtime *pgxpool.Pool,
	tenantID, actorID uuid.UUID,
) {
	t.Helper()
	importID, sourceJobID := mustRetryV7(t), mustRetryV7(t)
	seedFormsImportRetrySource(t, ctx, database, tenantID, importID,
		sourceJobID, actorID, mustRetryV7(t), "dead", false)
	call := formsImportRetryCall{
		tenantID: tenantID, importID: importID, sourceJobID: sourceJobID,
		retryJobID: mustRetryV7(t), receiptID: mustRetryV7(t), actorID: actorID,
		key: "stale-replay", requestSHA256: strings.Repeat("9", 64),
		createdAt: time.Date(2026, 8, 1, 12, 5, 0, 0, time.UTC),
	}
	result, err := invokeFormsImportRetry(ctx, runtime, call, false)
	if err != nil {
		t.Fatalf("create stale replay fixture: %v", err)
	}
	_, err = database.ExecContext(ctx,
		"UPDATE platform.jobs SET authority_group_generation=2 WHERE id=$1",
		result.jobID)
	if err == nil {
		_, err = database.ExecContext(ctx, `
UPDATE platform.idempotency_receipts
 SET authority_group_generation=2 WHERE receipt_id=$1`, result.receiptID)
	}
	if err != nil {
		t.Fatal(err)
	}
	call.retryJobID, call.receiptID = mustRetryV7(t), mustRetryV7(t)
	_, err = invokeFormsImportRetry(ctx, runtime, call, false)
	assertRetrySQLState(t, err, "42501")
}
