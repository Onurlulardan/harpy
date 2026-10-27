package domain

import "testing"

func TestNextStatusRequiresExplicitTerminalTransitions(t *testing.T) {
	if got, err := NextStatus(StatusOpen, StatusDone, TransitionComplete); err != nil || got != StatusDone {
		t.Fatalf("complete = %q, %v", got, err)
	}
	invalid := []struct {
		from, to TaskStatus
		kind     Transition
	}{
		{StatusOpen, StatusCancelled, TransitionMove},
		{StatusDone, StatusOpen, TransitionMove},
		{StatusDone, StatusInProgress, TransitionReopen},
	}
	for _, value := range invalid {
		if _, err := NextStatus(value.from, value.to, value.kind); err == nil {
			t.Fatalf("accepted transition %+v", value)
		}
	}
}

func TestTaskPriorityVocabularyMatchesCompatibilityContract(t *testing.T) {
	for _, priority := range []TaskPriority{
		PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent,
	} {
		if !priority.Valid() {
			t.Fatalf("priority %q rejected", priority)
		}
	}
	if TaskPriority("UNKNOWN").Valid() {
		t.Fatal("unknown priority accepted")
	}
}
