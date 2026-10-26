package postgres

type BoundExecutor struct {
	transaction *RuntimeTransaction
	state       *bindingState
}

func (executor BoundExecutor) usable() bool {
	return executor.state.usable(executor.transaction)
}
