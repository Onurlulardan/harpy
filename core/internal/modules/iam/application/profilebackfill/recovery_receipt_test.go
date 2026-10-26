package profilebackfill

import "testing"

func TestRecoveryReceiptAcceptsPersistedLifecycleStates(t *testing.T) {
	parent := recoveryFixtureParent()
	plan := recoveryFixturePlan(t, parent, 1)
	receipt := RecoveryReceipt{
		RequestSHA256:      parent.PlanSHA256,
		PlanID:             plan.ID(),
		RecoveryOfPlanID:   plan.RecoveryOfPlanID(),
		RecoveryRootPlanID: plan.RecoveryRootPlanID(),
		RecoverySequence:   plan.RecoverySequence(),
		Revision:           1,
		State:              RecoveryRunning,
		AuthorizationClock: 1,
	}
	if err := receipt.Validate(); err != nil {
		t.Fatalf("RecoveryReceipt.Validate() error = %v", err)
	}
	receipt.State = "unknown"
	if err := receipt.Validate(); err == nil {
		t.Fatal("RecoveryReceipt.Validate() accepted unknown state")
	}
}
