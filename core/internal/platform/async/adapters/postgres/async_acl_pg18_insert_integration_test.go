//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertACLInsertColumns(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	for _, check := range []struct{ table, allowed, denied string }{
		{"platform.jobs", "id", "created_at"},
		{"platform.job_attempts", "attempt_id", "finished_at"},
		{"platform.outbox", "event_id", "created_at"},
		{"platform.inbox", "id", "recorded_at"},
		{"platform.idempotency_receipts", "receipt_id", ""},
		{"platform.dead_letters", "id", ""},
	} {
		assertACLColumnPrivilege(t, ctx, database,
			check.table, check.allowed, "INSERT", true)
		if check.denied != "" {
			assertACLColumnPrivilege(t, ctx, database,
				check.table, check.denied, "INSERT", false)
		}
	}
}
