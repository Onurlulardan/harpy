//go:build integration

package v1adoption

import (
	"context"
	"database/sql"
)

type workSQLRows struct{ values *sql.Rows }

func (rows *workSQLRows) Next() bool               { return rows.values.Next() }
func (rows *workSQLRows) Scan(values ...any) error { return rows.values.Scan(values...) }
func (rows *workSQLRows) Err() error               { return rows.values.Err() }
func (rows *workSQLRows) Close()                   { _ = rows.values.Close() }

func workDatabaseAccess(transaction *sql.Tx) DatabaseAccess {
	return DatabaseAccess{
		Execute: func(ctx context.Context, statement string, arguments ...any) (int64, error) {
			result, err := transaction.ExecContext(ctx, statement, arguments...)
			if err != nil {
				return 0, err
			}
			return result.RowsAffected()
		},
		Query: func(ctx context.Context, statement string, arguments ...any) (Rows, error) {
			values, err := transaction.QueryContext(ctx, statement, arguments...)
			if err != nil {
				return nil, err
			}
			return &workSQLRows{values: values}, nil
		},
		QueryRow: func(ctx context.Context, statement string, arguments ...any) Row {
			return transaction.QueryRowContext(ctx, statement, arguments...)
		},
		CurrentSchema: func(ctx context.Context) (string, error) {
			var schema string
			err := transaction.QueryRowContext(ctx, "SELECT current_schema()").Scan(&schema)
			return schema, err
		},
	}
}

func beginWorkSchema(ctx context.Context, database *sql.DB, schema string) (*sql.Tx, error) {
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	_, err = transaction.ExecContext(ctx,
		"SELECT pg_catalog.set_config('search_path', $1, true)", schema+",pg_catalog")
	if err != nil {
		_ = transaction.Rollback()
		return nil, err
	}
	return transaction, nil
}
