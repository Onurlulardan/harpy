package postgres

import (
	"context"
	"sync"
)

type boundRow struct {
	executor  BoundExecutor
	context   context.Context
	statement string
	arguments []any
	once      sync.Once
	err       error
}

func newBoundRow(
	executor BoundExecutor,
	ctx context.Context,
	statement string,
	arguments []any,
) *boundRow {
	return &boundRow{
		executor: executor, context: ctx, statement: statement,
		arguments: append([]any(nil), arguments...),
	}
}

func (row *boundRow) Scan(destinations ...any) error {
	row.once.Do(func() {
		operation, err := row.executor.state.acquire(row.context, row.executor.transaction)
		if err != nil {
			row.err = err
			return
		}
		defer operation.finish()
		row.err = row.executor.transaction.transaction.QueryRow(
			operation.context, row.statement, row.arguments...,
		).Scan(destinations...)
	})
	return row.err
}
