package recoveryexecution

import (
	"strings"
	"testing"
)

func TestClaimNextRecoverySetUsesDurableAttemptRoutine(t *testing.T) {
	if !strings.Contains(claimNextRecoverySetSQL, "platformops_claim_next_recovery_set_v3") {
		t.Fatal("durable recovery claim routine is not bound")
	}
}
