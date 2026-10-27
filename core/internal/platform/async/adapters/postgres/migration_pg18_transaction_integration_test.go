//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func withAsyncRolledBackTransaction(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	test func(*sql.Tx),
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin isolated async test transaction: %v", err)
	}
	t.Cleanup(func() { _ = transaction.Rollback() })
	test(transaction)
}

func requireAsyncConstraintRejection(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("PostgreSQL accepted an invalid async row")
	}
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || !strings.HasPrefix(postgresError.Code, "23") {
		t.Fatalf("async row failed outside PostgreSQL constraint enforcement: %v", err)
	}
}

func requireAsyncNamedConstraint(t *testing.T, err error, name string) {
	t.Helper()
	requireAsyncConstraintRejection(t, err)
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.ConstraintName != name {
		t.Fatalf("constraint = %q, want %q", postgresError.ConstraintName, name)
	}
}
