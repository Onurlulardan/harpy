package invitation

type State string

const (
	StatePending  State = "pending"
	StateAccepted State = "accepted"
	StateRevoked  State = "revoked"
	StateExpired  State = "expired"
)

func (state State) Valid() bool {
	return state == StatePending || state == StateAccepted ||
		state == StateRevoked || state == StateExpired
}

func (state State) Terminal() bool { return state.Valid() && state != StatePending }
