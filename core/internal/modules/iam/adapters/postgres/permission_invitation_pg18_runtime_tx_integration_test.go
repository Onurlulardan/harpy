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

func iamPermissionInvitationID(prefix string, marker string) string {
	return fmt.Sprintf("%s000000-0000-4000-8000-0000000000%s", prefix, marker)
}

func beginIAMPGFirstRuntimeTx(
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
	want := "pg_catalog, " + schema
	var bound string
	if err := transaction.QueryRowContext(
		ctx, "SELECT pg_catalog.set_config('search_path', $1, true)", want,
	).Scan(&bound); err != nil {
		_ = transaction.Rollback()
		_ = connection.Close()
		return nil, nil, fmt.Errorf("bind PG-first schema %s: %w", schema, err)
	}
	if bound != want {
		_ = transaction.Rollback()
		_ = connection.Close()
		return nil, nil, fmt.Errorf("bind PG-first schema returned %q", bound)
	}
	return connection, transaction, nil
}

func mustRunIAMPGFirstText(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	statement string,
	arguments ...any,
) string {
	t.Helper()
	connection, transaction, err := beginIAMPGFirstRuntimeTx(ctx, database, schema)
	if err != nil {
		t.Fatalf("begin %s PG-first runtime transaction: %v", schema, err)
	}
	defer connection.Close()
	var result string
	if err := transaction.QueryRowContext(ctx, statement, arguments...).Scan(&result); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("run %s PG-first runtime query: %s", schema, iamPGErrorDetail(err))
	}
	if err := transaction.Commit(); err != nil {
		t.Fatalf("commit %s PG-first runtime query: %v", schema, err)
	}
	return result
}

func iamPGErrorDetail(err error) string {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		return fmt.Sprintf(
			"%s (SQLSTATE %s) detail=%q where=%q",
			postgresError.Message, postgresError.Code,
			postgresError.Detail, postgresError.Where,
		)
	}
	return err.Error()
}

func rejectIAMPGFirst(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	statement string,
	arguments ...any,
) error {
	t.Helper()
	connection, transaction, err := beginIAMPGFirstRuntimeTx(ctx, database, schema)
	if err != nil {
		t.Fatalf("begin %s rejected PG-first transaction: %v", schema, err)
	}
	defer connection.Close()
	_, err = transaction.ExecContext(ctx, statement, arguments...)
	_ = transaction.Rollback()
	if err == nil {
		t.Fatalf("PG-first IAM statement unexpectedly succeeded in %s", schema)
	}
	return err
}

func rollbackIAMPGFirstText(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	statement string,
	arguments ...any,
) {
	t.Helper()
	connection, transaction, err := beginIAMPGFirstRuntimeTx(ctx, database, schema)
	if err != nil {
		t.Fatalf("begin %s rollback probe: %v", schema, err)
	}
	defer connection.Close()
	var result string
	if err := transaction.QueryRowContext(ctx, statement, arguments...).Scan(&result); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("run %s rollback probe: %v", schema, err)
	}
	if err := transaction.Rollback(); err != nil {
		t.Fatalf("rollback %s probe: %v", schema, err)
	}
}
