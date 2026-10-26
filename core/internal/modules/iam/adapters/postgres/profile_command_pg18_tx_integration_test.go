//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type iamProfileCommandResult struct {
	id       uuid.UUID
	revision int64
	clock    int64
	replayed bool
}

func beginIAMProfileCommandRuntimeTx(
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
		if _, err = transaction.ExecContext(ctx, statement); err != nil {
			_ = transaction.Rollback()
			_ = connection.Close()
			return nil, nil, err
		}
	}
	var bound string
	err = transaction.QueryRowContext(
		ctx, "SELECT pg_catalog.set_config('search_path', $1, true)",
		"pg_catalog, "+schema,
	).Scan(&bound)
	if err != nil {
		_ = transaction.Rollback()
		_ = connection.Close()
		return nil, nil, fmt.Errorf("bind production IAM schema: %q: %w", bound, err)
	}
	if bound != "pg_catalog, "+schema {
		_ = transaction.Rollback()
		_ = connection.Close()
		return nil, nil, fmt.Errorf("bind production IAM schema returned %q", bound)
	}
	return connection, transaction, nil
}

func scanIAMProfileCommand(
	row *sql.Row,
) (iamProfileCommandResult, error) {
	var result iamProfileCommandResult
	err := row.Scan(&result.id, &result.revision, &result.clock, &result.replayed)
	return result, err
}

func finishIAMProfileCommand(
	connection *sql.Conn,
	transaction *sql.Tx,
	err error,
) error {
	defer connection.Close()
	if err != nil {
		_ = transaction.Rollback()
		return err
	}
	return transaction.Commit()
}
