//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
)

func beginIAMAtomicRuntimeTx(
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
	err = transaction.QueryRowContext(
		ctx, "SELECT pg_catalog.set_config('search_path', $1, true)",
		schema+", pg_catalog",
	).Scan(&bound)
	if err != nil {
		_ = transaction.Rollback()
		_ = connection.Close()
		return nil, nil, fmt.Errorf("bind runtime schema %s: %w", schema, err)
	}
	if bound != schema+", pg_catalog" {
		_ = transaction.Rollback()
		_ = connection.Close()
		return nil, nil, fmt.Errorf("bind runtime schema %s returned %q", schema, bound)
	}
	return connection, transaction, nil
}

func runIAMAtomicText(
	ctx context.Context,
	database *sql.DB,
	schema string,
	statement string,
	arguments ...any,
) (string, error) {
	connection, transaction, err := beginIAMAtomicRuntimeTx(ctx, database, schema)
	if err != nil {
		return "", err
	}
	defer connection.Close()
	var outcome string
	if err := transaction.QueryRowContext(ctx, statement, arguments...).Scan(&outcome); err != nil {
		_ = transaction.Rollback()
		return "", err
	}
	if err := transaction.Commit(); err != nil {
		return "", err
	}
	return outcome, nil
}

func mustRunIAMAtomicText(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	statement string,
	arguments ...any,
) string {
	t.Helper()
	outcome, err := runIAMAtomicText(ctx, database, schema, statement, arguments...)
	if err != nil {
		t.Fatalf("run %s IAM atomic routine: %v", schema, err)
	}
	return outcome
}

func mustRunIAMAtomicBigint(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	statement string,
	arguments ...any,
) int64 {
	t.Helper()
	connection, transaction, err := beginIAMAtomicRuntimeTx(ctx, database, schema)
	if err != nil {
		t.Fatalf("begin %s IAM bigint routine: %v", schema, err)
	}
	defer connection.Close()
	var result int64
	if err := transaction.QueryRowContext(ctx, statement, arguments...).Scan(&result); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("run %s IAM bigint routine: %v", schema, err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatalf("commit %s IAM bigint routine: %v", schema, err)
	}
	return result
}

func mustQueryIAMTenantValue(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	query string,
	destination ...any,
) {
	t.Helper()
	connection, transaction := beginIAMTenantTx(t, ctx, database, schema)
	defer connection.Close()
	if err := transaction.QueryRowContext(ctx, query).Scan(destination...); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("query %s IAM value: %v", schema, err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatalf("commit %s IAM query: %v", schema, err)
	}
}
