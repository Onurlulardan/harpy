package api

import "time"

type OnboardingOperation struct {
	plan      OnboardingPlan
	state     OnboardingState
	revision  int64
	updatedAt time.Time
}

func NewOnboardingOperation(
	plan OnboardingPlan,
	state OnboardingState,
	revision int64,
	updatedAt time.Time,
) (OnboardingOperation, error) {
	if !plan.IsValid() || !state.valid() || revision < 1 ||
		!microsecondUTC(updatedAt) || updatedAt.Before(plan.CreatedAt()) {
		return OnboardingOperation{}, ErrOnboardingInvalid
	}
	return OnboardingOperation{plan: plan, state: state,
		revision: revision, updatedAt: updatedAt}, nil
}

func (operation OnboardingOperation) Plan() OnboardingPlan   { return operation.plan }
func (operation OnboardingOperation) State() OnboardingState { return operation.state }
func (operation OnboardingOperation) Revision() int64        { return operation.revision }
func (operation OnboardingOperation) UpdatedAt() time.Time   { return operation.updatedAt }

func (operation OnboardingOperation) IsValid() bool {
	_, err := NewOnboardingOperation(
		operation.plan, operation.state, operation.revision, operation.updatedAt,
	)
	return err == nil
}
