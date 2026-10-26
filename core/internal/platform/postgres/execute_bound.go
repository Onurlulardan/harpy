package postgres

import "context"

func (executor BoundExecutor) Execute(
	ctx context.Context,
	statement string,
	arguments ...any,
) (CommandResult, error) {
	operation, err := executor.state.acquire(ctx, executor.transaction)
	if err != nil {
		return CommandResult{}, err
	}
	defer operation.finish()
	result, err := executor.transaction.transaction.Exec(
		operation.context, statement, arguments...,
	)
	if err != nil {
		return CommandResult{}, err
	}
	return CommandResult{rowsAffected: result.RowsAffected()}, nil
}
