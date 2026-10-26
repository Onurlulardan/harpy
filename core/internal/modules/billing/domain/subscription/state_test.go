package subscription

import "testing"

func TestStateVocabularyAndTerminalSet(t *testing.T) {
	states := []State{Pending, Trialing, Active, PastDue, Paused, Cancelled, Expired}
	for _, state := range states {
		if !state.Valid() {
			t.Fatalf("state %q is invalid", state)
		}
		wantTerminal := state == Cancelled || state == Expired
		if state.Terminal() != wantTerminal {
			t.Fatalf("state %q terminal = %t", state, state.Terminal())
		}
	}
	if State("unknown").Valid() || State("unknown").Terminal() {
		t.Fatal("unknown state was accepted")
	}
}
