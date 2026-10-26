//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func beginLayoutRuntimeTx(
	ctx context.Context,
	database *sql.DB,
	schema string,
) (*sql.Conn, *sql.Tx, error) {
	connection, err := database.Conn(ctx)
	if err != nil {
		return nil, nil, err
	}
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		_ = connection.Close()
		return nil, nil, err
	}
	for _, statement := range []string{
		"SET LOCAL SESSION AUTHORIZATION harpy_v2_runtime",
		"SET LOCAL ROLE harpy_runtime",
	} {
		if _, err := transaction.ExecContext(ctx, statement); err != nil {
			_ = transaction.Rollback()
			_ = connection.Close()
			return nil, nil, err
		}
	}
	var bound string
	err = transaction.QueryRowContext(ctx,
		"SELECT pg_catalog.set_config('search_path', $1, true)",
		"pg_catalog, "+schema).Scan(&bound)
	if err != nil || bound != "pg_catalog, "+schema {
		_ = transaction.Rollback()
		_ = connection.Close()
		return nil, nil, fmt.Errorf("bind layout runtime %s: %w", schema, err)
	}
	return connection, transaction, nil
}

func runLayoutRuntimeText(
	ctx context.Context,
	database *sql.DB,
	schema string,
	statement string,
	arguments ...any,
) (string, error) {
	connection, transaction, err := beginLayoutRuntimeTx(ctx, database, schema)
	if err != nil {
		return "", err
	}
	defer connection.Close()
	var result string
	if err := transaction.QueryRowContext(ctx, statement, arguments...).Scan(&result); err != nil {
		_ = transaction.Rollback()
		return "", err
	}
	if err := transaction.Commit(); err != nil {
		return "", err
	}
	return result, nil
}

func assertLayoutSQLState(t *testing.T, err error, expected ...string) {
	t.Helper()
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		t.Fatalf("layout rejection is not PostgreSQL error: %v", err)
	}
	for _, state := range expected {
		if postgresError.Code == state {
			return
		}
	}
	t.Fatalf("layout SQLSTATE = %s, want %v: %v", postgresError.Code, expected, err)
}
