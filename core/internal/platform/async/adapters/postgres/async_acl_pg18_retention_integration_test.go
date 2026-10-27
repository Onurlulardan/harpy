//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func disableAsyncOwnerLogin(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)
	if _, err := database.ExecContext(ctx, "ALTER ROLE harpy_owner NOLOGIN"); err != nil {
		t.Fatalf("disable async owner login: %v", err)
	}
	var canLogin bool
	if err := database.QueryRowContext(ctx, `
SELECT rolcanlogin FROM pg_catalog.pg_roles WHERE rolname='harpy_owner'`,
	).Scan(&canLogin); err != nil || canLogin {
		t.Fatalf("async owner login state = %t: %v", canLogin, err)
	}
}

func assertAsyncACLRetentionEntrypoints(
	t *testing.T,
	ctx context.Context,
	runtime *pgxpool.Conn,
	database *sql.DB,
) {
	t.Helper()
	var current, session string
	err := runtime.QueryRow(ctx, "SELECT current_user,session_user").
		Scan(&current, &session)
	if err != nil || current != "harpy_runtime" ||
		session != "harpy_v2_runtime" {
		t.Fatalf("retention runtime = %s/%s: %v", current, session, err)
	}
	counts, err := runRetentionCycleCounts(ctx, runtime)
	if err != nil || counts != [7]int32{} {
		t.Fatalf("retention cycle counts = %v: %v", counts, err)
	}
	for _, role := range []string{
		"harpy_v2_runtime", "harpy_adopter", "harpy_cluster_admin",
	} {
		assertACLFunctionPrivilege(
			t, ctx, database, role,
			"platform.platform_async_run_retention_cycle_v1()", false,
		)
		for _, routine := range []string{
			"platform.platform_async_redact_job_payloads_v1(integer)",
			"platform.platform_async_redact_outbox_payloads_v2(integer)",
			"platform.platform_async_redact_dead_letter_payloads_v2(integer)",
			"platform.platform_async_purge_inbox_v1(integer)",
			"platform.platform_async_purge_idempotency_receipts_v1(integer)",
			"platform.platform_async_purge_jobs_v1(integer)",
			"platform.platform_async_purge_dead_letters_v2(integer)",
		} {
			assertACLFunctionPrivilege(t, ctx, database, role, routine, false)
		}
	}
	assertMigratorRetentionDenied(t, ctx, database)
}

func assertMigratorRetentionDenied(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	pool := openAdoptionRole(
		t, ctx, database, "harpy_migrator", "migrator_integration_only",
	)
	connection, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Release()
	if _, err = connection.Exec(ctx, "SET ROLE harpy_owner"); err != nil {
		t.Fatal(err)
	}
	_, _, err = queryRetentionCycle(ctx, connection)
	var databaseError *pgconn.PgError
	if !errors.As(err, &databaseError) || databaseError.Code != "42501" {
		t.Fatalf("migrator retention cycle = %v", err)
	}
}
