package retention

import (
	"fmt"
	"time"
)

type CycleObservation struct {
	outcome  CycleOutcome
	duration time.Duration
	counts   StepCounts
}

func NewCycleObservation(
	outcome CycleOutcome,
	duration time.Duration,
	counts StepCounts,
) (CycleObservation, error) {
	if duration < 0 || (outcome != OutcomeWorked && outcome != OutcomeIdle &&
		outcome != OutcomeContended) {
		return CycleObservation{}, fmt.Errorf("%w: observation metadata", ErrResultInvalid)
	}
	result, err := NewCycleResult(outcome != OutcomeContended, counts)
	if err != nil || outcome == OutcomeWorked && !result.Worked() ||
		outcome == OutcomeIdle && result.Worked() {
		return CycleObservation{}, fmt.Errorf("%w: observation outcome", ErrResultInvalid)
	}
	return CycleObservation{outcome: outcome, duration: duration, counts: counts}, nil
}

func (value CycleObservation) Outcome() CycleOutcome   { return value.outcome }
func (value CycleObservation) Duration() time.Duration { return value.duration }
func (value CycleObservation) Counts() StepCounts      { return value.counts }
