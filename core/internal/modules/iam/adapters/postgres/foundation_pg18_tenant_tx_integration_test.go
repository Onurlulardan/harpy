//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func beginIAMTenantTx(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) (*sql.Conn, *sql.Tx) {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatalf("reserve %s IAM connection: %v", schema, err)
	}
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		_ = connection.Close()
		t.Fatalf("begin %s IAM transaction: %v", schema, err)
	}
	var bound string
	if err := transaction.QueryRowContext(
		ctx,
		"SELECT pg_catalog.set_config('search_path', $1, true)",
		schema+", pg_catalog",
	).Scan(&bound); err != nil {
		_ = transaction.Rollback()
		_ = connection.Close()
		t.Fatalf("bind %s IAM transaction: %v", schema, err)
	}
	if bound != schema+", pg_catalog" {
		_ = transaction.Rollback()
		_ = connection.Close()
		t.Fatalf("bound IAM search_path = %q", bound)
	}
	return connection, transaction
}

func mustExecIAMTenant(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	statement string,
	arguments ...any,
) {
	t.Helper()
	connection, transaction := beginIAMTenantTx(t, ctx, database, schema)
	defer connection.Close()
	if _, err := transaction.ExecContext(ctx, statement, arguments...); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("execute IAM statement in %s: %v", schema, err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatalf("commit IAM statement in %s: %v", schema, err)
	}
}

func rejectIAMTenant(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	statement string,
	arguments ...any,
) error {
	t.Helper()
	connection, transaction := beginIAMTenantTx(t, ctx, database, schema)
	defer connection.Close()
	_, err := transaction.ExecContext(ctx, statement, arguments...)
	_ = transaction.Rollback()
	if err == nil {
		t.Fatalf("IAM statement unexpectedly succeeded in %s", schema)
	}
	return err
}

func assertIAMSQLState(t *testing.T, err error, expected ...string) {
	t.Helper()
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		t.Fatalf("IAM rejection is not a PostgreSQL error: %v", err)
	}
	for _, code := range expected {
		if postgresError.Code == code {
			return
		}
	}
	t.Fatalf("IAM SQLSTATE = %s, want one of %v: %v", postgresError.Code, expected, err)
}

func assertIAMConstraint(
	t *testing.T,
	err error,
	code string,
	constraint string,
) {
	t.Helper()
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		t.Fatalf("IAM rejection is not a PostgreSQL error: %v", err)
	}
	if postgresError.Code != code || postgresError.ConstraintName != constraint {
		t.Fatalf(
			"IAM rejection = SQLSTATE %s constraint %q, want %s %q: %v",
			postgresError.Code,
			postgresError.ConstraintName,
			code,
			constraint,
			err,
		)
	}
}
