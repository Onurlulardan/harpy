package postgres

import "context"

func (executor BoundExecutor) Query(
	ctx context.Context,
	statement string,
	arguments ...any,
) (Rows, error) {
	operation, err := executor.state.acquire(ctx, executor.transaction)
	if err != nil {
		return nil, err
	}
	rows, err := executor.transaction.transaction.Query(
		operation.context, statement, arguments...,
	)
	if err != nil {
		operation.finish()
		return nil, err
	}
	return newLockedRows(rows, operation), nil
}
