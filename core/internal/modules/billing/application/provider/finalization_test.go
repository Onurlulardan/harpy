package providerapp

import "testing"

func TestFinalizeDecisionValidAcceptsOnlyClosedDecisionSet(t *testing.T) {
	accepted := []FinalizeDecision{
		FinalizePending, FinalizeSucceeded, FinalizeFailed,
		FinalizeReconcile, FinalizeReconcileUnknown, FinalizeReplay,
	}
	for _, decision := range accepted {
		if !decision.Valid() {
			t.Fatalf("decision %q rejected", decision)
		}
	}
	for _, decision := range []FinalizeDecision{"", "success", "unknown", "pending "} {
		if decision.Valid() {
			t.Fatalf("decision %q accepted", decision)
		}
	}
}

func TestFinalizeDecisionRequiresExactLedgerState(t *testing.T) {
	accepted := map[FinalizeDecision][]CommandState{
		FinalizePending:          {CommandQueued, CommandRunning, CommandReconcileRequired},
		FinalizeSucceeded:        {CommandSucceeded},
		FinalizeFailed:           {CommandFailed},
		FinalizeReconcile:        {CommandReconcileRequired},
		FinalizeReconcileUnknown: {CommandReconcileRequired},
		FinalizeReplay:           {CommandSucceeded, CommandFailed},
	}
	allStates := []CommandState{
		CommandQueued, CommandRunning, CommandSucceeded,
		CommandFailed, CommandReconcileRequired,
	}
	for decision, states := range accepted {
		for _, state := range allStates {
			want := false
			for _, acceptedState := range states {
				want = want || state == acceptedState
			}
			if got := decision.validState(state); got != want {
				t.Fatalf("decision %q state %q = %v, want %v", decision, state, got, want)
			}
		}
	}
	if FinalizeDecision("unknown").validState(CommandSucceeded) {
		t.Fatal("unknown decision accepted a ledger state")
	}
}
