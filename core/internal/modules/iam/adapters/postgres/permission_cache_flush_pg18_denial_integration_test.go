//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertPermissionCacheFlushDenied(
	t *testing.T, ctx context.Context, database *sql.DB,
	operationID string, request []byte, actor string,
	tokenVersion, authorizationVersion, platformClock int64, sqlState string,
) {
	t.Helper()
	connection, transaction := beginPermissionCacheRuntimeTx(t, ctx, database)
	defer connection.Close()
	var clock, revision int64
	var replayed bool
	err := transaction.QueryRowContext(ctx, permissionCacheFlushCall,
		operationID, request, actor, tokenVersion, authorizationVersion, platformClock,
	).Scan(&clock, &revision, &replayed)
	_ = transaction.Rollback()
	if err == nil {
		t.Fatalf("permission cache denial unexpectedly returned %d/%d/%t",
			clock, revision, replayed)
	}
	assertIAMSQLState(t, err, sqlState)
}

func beginPermissionCacheRuntimeTx(
	t *testing.T, ctx context.Context, database *sql.DB,
) (*sql.Conn, *sql.Tx) {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		_ = connection.Close()
		t.Fatal(err)
	}
	for _, statement := range []string{
		"SET LOCAL SESSION AUTHORIZATION harpy_v2_runtime",
		"SET LOCAL ROLE harpy_runtime",
	} {
		if _, err := transaction.ExecContext(ctx, statement); err != nil {
			_ = transaction.Rollback()
			_ = connection.Close()
			t.Fatal(err)
		}
	}
	var bound string
	if err := transaction.QueryRowContext(ctx,
		"SELECT pg_catalog.set_config('search_path',$1,true)",
		"pg_catalog,t_iam_a").Scan(&bound); err != nil || bound != "pg_catalog,t_iam_a" {
		_ = transaction.Rollback()
		_ = connection.Close()
		t.Fatalf("bind permission-cache runtime: %q %v", bound, err)
	}
	return connection, transaction
}
