package usage

type ReservationState string

const (
	Reserved          ReservationState = "reserved"
	Committed         ReservationState = "committed"
	Released          ReservationState = "released"
	ReconcileRequired ReservationState = "reconcile_required"
)

func (state ReservationState) Valid() bool {
	switch state {
	case Reserved, Committed, Released, ReconcileRequired:
		return true
	default:
		return false
	}
}
