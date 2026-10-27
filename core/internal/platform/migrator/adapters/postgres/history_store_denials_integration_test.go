//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/jackc/pgx/v5/pgconn"
	goosedb "github.com/pressly/goose/v3/database"
)

const insertTamperedHistorySQL = `INSERT INTO platform.platform_migration_history (
	version, record_kind, is_applied, owner, scope, origin,
	selected_path, selected_sha256, run_id,
	target_database_generation, next_minimum_runtime_generation, applied_at
) VALUES ($1, 'migration', TRUE, $2, $3, $4, $5, $6, $7, $8, $9,
	statement_timestamp())`

func assertHistoryWriteDenials(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	store *Store,
	selected []postgresv1.SelectedMigration,
) {
	t.Helper()
	version := selected[0].Version
	err := store.Insert(ctx, database, goosedb.InsertRequest{Version: version})
	assertPostgreSQLState(t, err, "23505", "duplicate version")

	unmanifestedVersion := selected[1].Version + 100
	err = store.Insert(ctx, database, goosedb.InsertRequest{
		Version: unmanifestedVersion,
	})
	if err == nil || !strings.Contains(err.Error(), "no validated history metadata") {
		t.Fatalf("unmanifested migration denial = %v", err)
	}

	_, err = database.ExecContext(ctx, insertTamperedHistorySQL,
		unmanifestedVersion, "platform_migrator", "platform", "fresh",
		"migrations/platform/20260714000300_tampered.fresh.sql",
		strings.Repeat("A", 64), testRunID, int64(3), int64(1),
	)
	assertPostgreSQLState(t, err, "23514", "tampered digest")

	_, err = database.ExecContext(ctx,
		"UPDATE platform.platform_migration_history SET owner = $1 WHERE version = $2",
		"attacker", version,
	)
	assertPostgreSQLState(t, err, "55000", "immutable update")

	_, err = database.ExecContext(ctx,
		"DELETE FROM platform.platform_migration_history WHERE version = $1",
		version,
	)
	assertPostgreSQLState(t, err, "55000", "immutable delete")

	err = store.Delete(ctx, database, version)
	if !errors.Is(err, ErrDownForbidden) {
		t.Fatalf("store delete error = %v, want ErrDownForbidden", err)
	}
	var downError *DownForbiddenError
	if !errors.As(err, &downError) || downError.Version != version {
		t.Fatalf("typed store delete error = %#v, want version %d", downError, version)
	}
}

func assertPostgreSQLState(
	t *testing.T,
	err error,
	wantCode string,
	action string,
) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s unexpectedly succeeded", action)
	}
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		t.Fatalf("%s error type = %T, want PostgreSQL error: %v", action, err, err)
	}
	if postgresError.Code != wantCode {
		t.Fatalf("%s SQLSTATE = %s, want %s: %v",
			action, postgresError.Code, wantCode, postgresError)
	}
}
