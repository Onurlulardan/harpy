package subscription

func ValidateTransition(from, to State) error {
	if !from.Valid() || !to.Valid() || from == to {
		return ErrTransition
	}
	allowed := false
	switch from {
	case Pending:
		allowed = to == Trialing || to == Active || to == Cancelled
	case Trialing:
		allowed = to == Active || to == PastDue || to == Cancelled || to == Expired
	case Active:
		allowed = to == PastDue || to == Paused || to == Cancelled || to == Expired
	case PastDue:
		allowed = to == Active || to == Paused || to == Cancelled || to == Expired
	case Paused:
		allowed = to == Active || to == Cancelled || to == Expired
	}
	if !allowed {
		return ErrTransition
	}
	return nil
}
