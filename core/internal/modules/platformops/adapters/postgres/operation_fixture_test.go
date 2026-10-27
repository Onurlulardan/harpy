package postgres

import (
	"context"
	"errors"
)

type operationRow func(...any) error

func (row operationRow) Scan(destinations ...any) error { return row(destinations...) }

type operationRows struct {
	rows  []func(...any) error
	index int
}

func (rows *operationRows) Close()     {}
func (rows *operationRows) Err() error { return nil }
func (rows *operationRows) Next() bool { return rows.index < len(rows.rows) }
func (rows *operationRows) Scan(destinations ...any) error {
	if rows.index >= len(rows.rows) {
		return errors.New("operation row is unavailable")
	}
	scan := rows.rows[rows.index]
	rows.index++
	return scan(destinations...)
}

type operationDatabase struct {
	row        operationRow
	rows       databaseRows
	executions int
}

func (database *operationDatabase) Execute(
	context.Context, string, ...any,
) (commandResult, error) {
	database.executions++
	return commandResult{}, nil
}

func (database *operationDatabase) Query(
	context.Context, string, ...any,
) (databaseRows, error) {
	return database.rows, nil
}

func (database *operationDatabase) QueryRow(
	context.Context, string, ...any,
) databaseRow {
	return database.row
}

func operationStore(database *operationDatabase) *Store {
	return &Store{transactions: func(
		ctx context.Context, _ transactionMode, callback func(boundDatabase) error,
	) error {
		return callback(database)
	}, rowNotFound: func(error) bool { return false }}
}
