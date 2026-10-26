package usage

import "testing"

func TestReservationTransitionMatrix(t *testing.T) {
	states := []ReservationState{Reserved, Committed, Released, ReconcileRequired}
	allowed := map[[2]ReservationState]bool{
		{Reserved, Committed}:          true,
		{Reserved, Released}:           true,
		{Reserved, ReconcileRequired}:  true,
		{ReconcileRequired, Committed}: true,
		{ReconcileRequired, Released}:  true,
	}
	for _, current := range states {
		for _, next := range states {
			_, err := Transition(current, next)
			if (err == nil) != allowed[[2]ReservationState{current, next}] {
				t.Fatalf("transition %s -> %s mismatch: %v", current, next, err)
			}
		}
	}
}

func TestReservationTransitionRejectsUnknownState(t *testing.T) {
	if _, err := Transition("unknown", Released); err != ErrTransition {
		t.Fatalf("expected ErrTransition, got %v", err)
	}
}
