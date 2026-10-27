package postgres

import "context"

func (executor BoundExecutor) QueryRow(
	ctx context.Context,
	statement string,
	arguments ...any,
) Row {
	return newBoundRow(executor, ctx, statement, arguments)
}
