package api

import "time"

type OnboardingTransition struct {
	OperationID UUID
	State       OnboardingState
	Revision    int64
	UpdatedAt   time.Time
}

func (transition OnboardingTransition) IsValid() bool {
	return transition.OperationID != NilUUID && transition.State.valid() &&
		transition.Revision > 0 && microsecondUTC(transition.UpdatedAt)
}
