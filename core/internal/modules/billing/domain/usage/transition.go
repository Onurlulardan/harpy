package usage

func CanTransition(current, next ReservationState) bool {
	switch current {
	case Reserved:
		return next == Committed || next == Released || next == ReconcileRequired
	case ReconcileRequired:
		return next == Committed || next == Released
	default:
		return false
	}
}

func Transition(current, next ReservationState) (ReservationState, error) {
	if !current.Valid() || !next.Valid() || !CanTransition(current, next) {
		return "", ErrTransition
	}
	return next, nil
}
