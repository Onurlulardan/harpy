//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func assertFormsImportRetryConcurrency(
	t *testing.T, ctx context.Context, database *sql.DB, runtime *pgxpool.Pool,
	tenantID, actorID uuid.UUID,
) {
	t.Helper()
	importID, sourceJobID := mustRetryV7(t), mustRetryV7(t)
	seedFormsImportRetrySource(t, ctx, database, tenantID, importID,
		sourceJobID, actorID, mustRetryV7(t), "dead", false)
	var calls [2]formsImportRetryCall
	for index := range calls {
		calls[index] = formsImportRetryCall{
			tenantID: tenantID, importID: importID, sourceJobID: sourceJobID,
			retryJobID: mustRetryV7(t), receiptID: mustRetryV7(t), actorID: actorID,
			key:           fmt.Sprintf("concurrent-%d", index),
			requestSHA256: strings.Repeat(fmt.Sprintf("%d", index+1), 64),
			createdAt:     time.Date(2026, 8, 1, 12, 3, 0, 0, time.UTC),
		}
	}
	start := make(chan struct{})
	var wait sync.WaitGroup
	var failures [2]error
	for index := range calls {
		wait.Add(1)
		go func(slot int) {
			defer wait.Done()
			<-start
			_, failures[slot] = invokeFormsImportRetry(ctx, runtime, calls[slot], false)
		}(index)
	}
	close(start)
	wait.Wait()
	successes, conflicts := 0, 0
	for _, err := range failures {
		if err == nil {
			successes++
			continue
		}
		var databaseError *pgconn.PgError
		if errors.As(err, &databaseError) && databaseError.Code == "P6045" {
			conflicts++
			continue
		}
		t.Fatalf("concurrent retry error = %v", err)
	}
	var jobs, receipts int
	err := database.QueryRowContext(ctx, `
SELECT count(*),count(receipt.receipt_id)
FROM platform.jobs AS job
LEFT JOIN platform.idempotency_receipts AS receipt
 ON receipt.async_job_id=job.id
 AND receipt.operation_id='forms.records.import.resume'
WHERE job.retry_of_job_id=$1`, sourceJobID).Scan(&jobs, &receipts)
	if err != nil || successes != 1 || conflicts != 1 || jobs != 1 || receipts != 1 {
		t.Fatalf("concurrent retry = %d/%d jobs=%d receipts=%d, %v",
			successes, conflicts, jobs, receipts, err)
	}
}
