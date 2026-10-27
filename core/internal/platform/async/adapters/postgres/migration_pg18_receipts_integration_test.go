//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertAsyncReceiptMatrix(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	t.Run("HTTP receipt leaves effect identity optional", func(t *testing.T) {
		withAsyncRolledBackTransaction(t, ctx, database, func(tx *sql.Tx) {
			id := "84000000-0000-4000-8000-000000000001"
			if err := insertAsyncHTTPReceipt(
				ctx, tx, id, asyncTenantA, "tenant", asyncTenantA,
				"interactive_write", nil,
			); err != nil {
				t.Fatalf("valid HTTP receipt rejected: %v", err)
			}
			var effectKind, provider sql.NullString
			if err := tx.QueryRowContext(ctx, `
SELECT effect_kind, provider
FROM platform.idempotency_receipts
WHERE receipt_id = $1`, id).Scan(&effectKind, &provider); err != nil {
				t.Fatalf("read optional effect identity: %v", err)
			}
			if effectKind.Valid || provider.Valid {
				t.Fatal("HTTP receipt populated effect-only identity")
			}
		})
	})
	t.Run("async handle references same-scope job", func(t *testing.T) {
		withAsyncRolledBackTransaction(t, ctx, database, func(tx *sql.Tx) {
			jobID := "84000000-0000-4000-8000-000000000002"
			if err := insertAsyncQueuedJob(
				ctx, tx, jobID, asyncTenantA, "tenant", asyncTenantA,
				"operations", "workflow", "receipt-job",
			); err != nil {
				t.Fatalf("insert referenced job: %v", err)
			}
			if err := insertAsyncHTTPReceipt(
				ctx, tx, "84000000-0000-4000-8000-000000000003",
				asyncTenantA, "tenant", asyncTenantA, "async_handle", jobID,
			); err != nil {
				t.Fatalf("same-scope async handle rejected: %v", err)
			}
		})
	})
	t.Run("async handle rejects cross-scope job", func(t *testing.T) {
		withAsyncRolledBackTransaction(t, ctx, database, func(tx *sql.Tx) {
			jobID := "84000000-0000-4000-8000-000000000004"
			if err := insertAsyncQueuedJob(
				ctx, tx, jobID, asyncTenantA, "tenant", asyncTenantA,
				"operations", "workflow", "cross-scope-job",
			); err != nil {
				t.Fatalf("insert tenant A job: %v", err)
			}
			err := insertAsyncHTTPReceipt(
				ctx, tx, "84000000-0000-4000-8000-000000000005",
				asyncTenantB, "tenant", asyncTenantB, "async_handle", jobID,
			)
			requireAsyncNamedConstraint(t, err, "idempotency_receipts_async_job_fk")
		})
	})
	assertAsyncReceiptShapeNegatives(t, ctx, database)
}

func assertAsyncReceiptShapeNegatives(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	t.Run("async handle requires job", func(t *testing.T) {
		withAsyncRolledBackTransaction(t, ctx, database, func(tx *sql.Tx) {
			err := insertAsyncHTTPReceipt(
				ctx, tx, "84000000-0000-4000-8000-000000000006",
				asyncTenantA, "tenant", asyncTenantA, "async_handle", nil,
			)
			requireAsyncConstraintRejection(t, err)
		})
	})
	t.Run("interactive receipt rejects job", func(t *testing.T) {
		withAsyncRolledBackTransaction(t, ctx, database, func(tx *sql.Tx) {
			jobID := "84000000-0000-4000-8000-000000000007"
			if err := insertAsyncQueuedJob(
				ctx, tx, jobID, asyncTenantA, "tenant", asyncTenantA,
				"operations", "workflow", "interactive-job",
			); err != nil {
				t.Fatalf("insert job: %v", err)
			}
			err := insertAsyncHTTPReceipt(
				ctx, tx, "84000000-0000-4000-8000-000000000008",
				asyncTenantA, "tenant", asyncTenantA, "interactive_write", jobID,
			)
			requireAsyncConstraintRejection(t, err)
		})
	})
	t.Run("effect requires kind and provider", func(t *testing.T) {
		withAsyncRolledBackTransaction(t, ctx, database, func(tx *sql.Tx) {
			err := insertAsyncEffectReceipt(
				ctx, tx, "84000000-0000-4000-8000-000000000009",
				asyncTenantA, "tenant", asyncTenantA, nil, nil,
			)
			requireAsyncConstraintRejection(t, err)
		})
	})
}
