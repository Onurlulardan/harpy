package retention

import (
	"errors"
	"testing"
)

func TestCycleResultValidatesSealedStepCounts(t *testing.T) {
	counts := StepCounts{
		JobsRedacted: 1, OutboxRedacted: 2, DeadLettersRedacted: 3,
		InboxPurged: 4, IdempotencyPurged: 5, JobsPurged: 6,
		DeadLettersPurged: 7,
	}
	result, err := NewCycleResult(true, counts)
	if err != nil || !result.LockAcquired() || !result.Worked() || result.Counts() != counts {
		t.Fatalf("valid cycle result = %#v, %v", result, err)
	}
	for name, candidate := range map[string]StepCounts{
		"negative":   {JobsRedacted: -1},
		"over batch": {DeadLettersPurged: 501},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := NewCycleResult(true, candidate); !errors.Is(err, ErrResultInvalid) {
				t.Fatalf("invalid count error = %v", err)
			}
		})
	}
	if _, err := NewCycleResult(false, StepCounts{JobsPurged: 1}); !errors.Is(err, ErrResultInvalid) {
		t.Fatalf("contended mutation error = %v", err)
	}
}
