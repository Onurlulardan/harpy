//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const retentionCycleCall = `SELECT
 lock_acquired,jobs_redacted,outbox_redacted,dead_letters_redacted,
 inbox_purged,idempotency_purged,jobs_purged,dead_letters_purged
FROM platform.platform_async_run_retention_cycle_v1()`

func TestAsyncRetentionCycleAuthorityOnPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startAdoptionPG18(t, ctx, "retention_cycle_"+origin)
			applyAsyncACLFixture(t, ctx, database, origin)
			disableAsyncOwnerLogin(t, ctx, database)
			runtime := openAsyncRuntimeRole(t, ctx, database)
			assertRetentionCycleZero(t, ctx, runtime)
			assertRetentionLeafPrivilegesRevoked(t, ctx, database)
			assertRetentionCycleAdvisoryLock(t, ctx, database, runtime)
		})
	}
}

type retentionCycleQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func assertRetentionCycleZero(
	t *testing.T, ctx context.Context, querier retentionCycleQuerier,
) bool {
	t.Helper()
	var acquired bool
	var counts [7]int32
	err := querier.QueryRow(ctx, retentionCycleCall).Scan(
		&acquired, &counts[0], &counts[1], &counts[2], &counts[3],
		&counts[4], &counts[5], &counts[6],
	)
	if err != nil {
		t.Fatalf("run retention cycle: %v", err)
	}
	for index, count := range counts {
		if count != 0 {
			t.Fatalf("retention count %d = %d, want zero", index, count)
		}
	}
	return acquired
}

func assertRetentionLeafPrivilegesRevoked(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	for _, routine := range []string{
		"platform.platform_async_redact_job_payloads_v1(integer)",
		"platform.platform_async_redact_outbox_payloads_v2(integer)",
		"platform.platform_async_redact_dead_letter_payloads_v2(integer)",
		"platform.platform_async_purge_inbox_v1(integer)",
		"platform.platform_async_purge_idempotency_receipts_v1(integer)",
		"platform.platform_async_purge_jobs_v1(integer)",
		"platform.platform_async_purge_dead_letters_v2(integer)",
	} {
		assertACLFunctionPrivilege(t, ctx, database, "harpy_runtime", routine, false)
	}
}

func assertRetentionCycleAdvisoryLock(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	first *pgxpool.Conn,
) {
	t.Helper()
	transaction, err := first.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()
	if !assertRetentionCycleZero(t, ctx, transaction) {
		t.Fatal("first retention cycle did not acquire its transaction lock")
	}
	second := openAsyncRuntimeRole(t, ctx, database)
	if assertRetentionCycleZero(t, ctx, second) {
		t.Fatal("concurrent retention cycle acquired the same transaction lock")
	}
}
