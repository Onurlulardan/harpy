//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestAsyncSliceAMigrationsAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startAsyncMigrationDatabase(t, ctx, origin)
			applyAsyncSliceA(t, ctx, database, origin)
			t.Run("catalog", func(t *testing.T) {
				assertAsyncSliceACatalog(t, ctx, database)
			})
			t.Run("tenant_global", func(t *testing.T) {
				assertAsyncTenantGlobalNegatives(t, ctx, database)
			})
			t.Run("outbox", func(t *testing.T) {
				assertAsyncOutboxMatrix(t, ctx, database)
			})
			t.Run("inbox", func(t *testing.T) {
				assertAsyncInboxMatrix(t, ctx, database)
			})
			t.Run("jobs", func(t *testing.T) {
				assertAsyncJobMatrices(t, ctx, database)
			})
			t.Run("receipts", func(t *testing.T) {
				assertAsyncReceiptMatrix(t, ctx, database)
			})
			t.Run("rollback", func(t *testing.T) {
				assertAsyncTransactionRollback(t, ctx, database)
			})
			t.Run("writer_sql", func(t *testing.T) {
				assertAsyncWriterSQL(t, ctx, database)
			})
		})
	}
}
