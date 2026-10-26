//go:build integration

package integrationtest

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func expectRecoveryRuntimeError(
	t *testing.T, ctx context.Context, db publicationDatabase,
	setRole bool, statement, expectedCode string,
) {
	t.Helper()
	conn := openRecoveryRuntimeConnection(t, ctx, db)
	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if setRole {
		_, err = tx.Exec(ctx, "SET LOCAL ROLE harpy_runtime")
		if err == nil {
			_, err = tx.Exec(ctx,
				"SELECT pg_catalog.set_config('search_path',$1,true)",
				db.tenant.schema+",platform,pg_catalog")
		}
	}
	if err == nil {
		_, err = tx.Exec(ctx, statement)
	}
	_ = tx.Rollback(ctx)
	assertRecoveryPGCode(t, err, expectedCode)
}

func expectRecoveryOwnerError(
	t *testing.T, ctx context.Context, db publicationDatabase,
	statement, expectedCode string, arguments ...any,
) {
	t.Helper()
	tx, err := db.admin.Begin(ctx)
	if err == nil {
		_, err = tx.Exec(ctx, "SET LOCAL ROLE harpy_owner")
	}
	if err == nil {
		_, err = tx.Exec(ctx,
			"SELECT pg_catalog.set_config('search_path',$1,true)",
			db.tenant.schema+",platform,pg_catalog")
	}
	if err == nil {
		arguments = append([]any{pgx.QueryExecModeSimpleProtocol}, arguments...)
		_, err = tx.Exec(ctx, statement, arguments...)
	}
	if err == nil {
		err = tx.Commit(ctx)
	} else {
		_ = tx.Rollback(ctx)
	}
	assertRecoveryPGCode(t, err, expectedCode)
}

func assertRecoveryPGCode(t *testing.T, err error, expectedCode string) {
	t.Helper()
	var databaseError *pgconn.PgError
	if !errors.As(err, &databaseError) || databaseError.Code != expectedCode {
		t.Fatalf("PostgreSQL error = %v, want SQLSTATE %s", err, expectedCode)
	}
}
