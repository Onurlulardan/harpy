package profilebackfill

import (
	"errors"
	"testing"
)

func TestStartRecoveryRejectsMissingDependencies(t *testing.T) {
	if _, err := NewRecoveryService(nil, nil); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("NewRecoveryService() error = %v", err)
	}
}

func TestStartRecoveryRequiresAtomicZeroActivation(t *testing.T) {
	parent := recoveryFixtureParent()
	zero := recoveryFixturePlan(t, parent, 0)
	result := RecoveryResult{
		PlanID: zero.ID(), Revision: 1, AuthorizationClock: 1,
		State: RecoveryActivated,
	}
	if !validRecoveryResult(result, zero) {
		t.Fatal("zero recovery activation rejected")
	}
	result.State = RecoveryPlanned
	if validRecoveryResult(result, zero) {
		t.Fatal("zero recovery accepted without atomic activation")
	}
}

func TestStartRecoveryAcceptsAdvancedReplayState(t *testing.T) {
	parent := recoveryFixtureParent()
	plan := recoveryFixturePlan(t, parent, 1)
	result := RecoveryResult{
		PlanID: plan.ID(), Revision: 3, AuthorizationClock: 1,
		State: RecoveryRunning, Replayed: true,
	}
	if !validRecoveryResult(result, plan) {
		t.Fatal("advanced recovery replay state rejected")
	}
	result.Replayed = false
	if validRecoveryResult(result, plan) {
		t.Fatal("new recovery accepted with advanced state")
	}
}
