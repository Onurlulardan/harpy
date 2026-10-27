package dashboardexport

type State string

const (
	StateQueued    State = "queued"
	StateRunning   State = "running"
	StateReady     State = "ready"
	StateFailed    State = "failed"
	StateCancelled State = "cancelled"
	StateDeleting  State = "deleting"
	StateDeleted   State = "deleted"
	StateExpired   State = "expired"
)

func (state State) Valid() bool {
	switch state {
	case StateQueued, StateRunning, StateReady, StateFailed, StateCancelled,
		StateDeleting, StateDeleted, StateExpired:
		return true
	default:
		return false
	}
}

func (state State) Terminal() bool {
	return state == StateReady || state == StateFailed || state == StateCancelled ||
		state == StateDeleted || state == StateExpired
}
