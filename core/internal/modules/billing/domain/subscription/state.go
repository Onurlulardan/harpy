package subscription

type State string

const (
	Pending   State = "pending"
	Trialing  State = "trialing"
	Active    State = "active"
	PastDue   State = "past_due"
	Paused    State = "paused"
	Cancelled State = "cancelled"
	Expired   State = "expired"
)

func (state State) Valid() bool {
	switch state {
	case Pending, Trialing, Active, PastDue, Paused, Cancelled, Expired:
		return true
	default:
		return false
	}
}

func (state State) Terminal() bool {
	return state == Cancelled || state == Expired
}

func (state State) purchasedGrantBearing() bool {
	return state == Trialing || state == Active || state == PastDue
}
