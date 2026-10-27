package lifecycle

func (value Operation) Valid() bool {
	if !value.ID.Valid() || !validKind(value.Kind) || !validState(value.State) ||
		!value.Plan.Valid() || !canonicalTime(value.StartedAt) || value.Revision < 1 {
		return false
	}
	if value.RetryOf != nil && (!value.RetryOf.Valid() || *value.RetryOf == value.ID) {
		return false
	}
	if value.DestructiveStarted != nil && (!canonicalTime(*value.DestructiveStarted) ||
		value.Kind != OperationPurge || value.DestructiveStarted.Before(value.StartedAt)) {
		return false
	}
	return true
}

func validState(value OperationState) bool {
	switch value {
	case OperationPlanned, OperationApplying, OperationWaiting, OperationVerifying,
		OperationActivating, OperationCancelling, OperationSucceeded, OperationBlocked,
		OperationFailed, OperationCancelled:
		return true
	default:
		return false
	}
}
