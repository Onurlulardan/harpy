//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertAsyncTransactionRollback(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin async rollback proof: %v", err)
	}
	jobID := "85000000-0000-4000-8000-000000000001"
	if err := insertAsyncQueuedJob(
		ctx, transaction, jobID, asyncTenantA, "tenant", asyncTenantA,
		"operations", "workflow", "rollback-job",
	); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("insert rollback job: %v", err)
	}
	outbox := newAsyncOutboxRow("85000000-0000-4000-8000-000000000002")
	if err := insertAsyncOutbox(ctx, transaction, outbox); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("insert rollback outbox: %v", err)
	}
	if err := insertAsyncHTTPReceipt(
		ctx, transaction, "85000000-0000-4000-8000-000000000003",
		asyncTenantA, "tenant", asyncTenantA, "interactive_write", nil,
	); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("insert rollback receipt: %v", err)
	}
	if err := transaction.Rollback(); err != nil {
		t.Fatalf("rollback async transaction: %v", err)
	}
	for _, table := range []string{"jobs", "outbox", "idempotency_receipts"} {
		var count int
		query := "SELECT count(*) FROM platform." + table
		if err := database.QueryRowContext(ctx, query).Scan(&count); err != nil {
			t.Fatalf("count rolled back %s: %v", table, err)
		}
		if count != 0 {
			t.Fatalf("rollback retained %d %s rows", count, table)
		}
	}
}
