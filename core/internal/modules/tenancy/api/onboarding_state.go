package api

type OnboardingState string

const (
	OnboardingAccepted              OnboardingState = "accepted"
	OnboardingMigrating             OnboardingState = "migrating"
	OnboardingMigrationReady        OnboardingState = "migration_ready"
	OnboardingAuthorityPending      OnboardingState = "authority_pending"
	OnboardingAuthorityAcknowledged OnboardingState = "authority_acknowledged"
	OnboardingFinalizing            OnboardingState = "finalizing"
	OnboardingSucceeded             OnboardingState = "succeeded"
	OnboardingFailed                OnboardingState = "failed"
	OnboardingFenced                OnboardingState = "fenced"
)

func (state OnboardingState) valid() bool {
	switch state {
	case OnboardingAccepted, OnboardingMigrating, OnboardingMigrationReady,
		OnboardingAuthorityPending, OnboardingAuthorityAcknowledged,
		OnboardingFinalizing, OnboardingSucceeded, OnboardingFailed,
		OnboardingFenced:
		return true
	default:
		return false
	}
}
