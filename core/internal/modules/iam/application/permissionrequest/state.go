package permissionrequest

type State string

const (
	StatePending   State = "PENDING"
	StateInReview  State = "IN_REVIEW"
	StateApproved  State = "APPROVED"
	StateRejected  State = "REJECTED"
	StateCancelled State = "CANCELLED"
	StateExpired   State = "EXPIRED"
)

func (state State) Valid() bool {
	switch state {
	case StatePending, StateInReview, StateApproved, StateRejected,
		StateCancelled, StateExpired:
		return true
	default:
		return false
	}
}

func (state State) Terminal() bool {
	return state == StateApproved || state == StateRejected ||
		state == StateCancelled || state == StateExpired
}

func CanTransition(from, to State) bool {
	switch from {
	case StatePending:
		return to == StateInReview || to == StateCancelled || to == StateExpired
	case StateInReview:
		return to == StateApproved || to == StateRejected ||
			to == StateCancelled || to == StateExpired
	default:
		return false
	}
}
