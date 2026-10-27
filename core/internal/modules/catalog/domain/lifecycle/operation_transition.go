package lifecycle

func operationTransition(current, next OperationState) bool {
	allowed := map[OperationState]map[OperationState]bool{
		OperationPlanned: {
			OperationApplying: true, OperationCancelled: true,
			OperationBlocked: true, OperationFailed: true,
		},
		OperationApplying: {
			OperationWaiting: true, OperationVerifying: true,
			OperationCancelling: true, OperationBlocked: true, OperationFailed: true,
		},
		OperationWaiting: {
			OperationApplying: true, OperationVerifying: true,
			OperationCancelling: true, OperationBlocked: true, OperationFailed: true,
		},
		OperationVerifying: {
			OperationActivating: true, OperationCancelling: true,
			OperationBlocked: true, OperationFailed: true,
		},
		OperationActivating: {OperationSucceeded: true, OperationFailed: true},
		OperationCancelling: {OperationCancelled: true, OperationBlocked: true},
		OperationBlocked:    {OperationApplying: true, OperationCancelling: true},
	}
	return allowed[current][next]
}
