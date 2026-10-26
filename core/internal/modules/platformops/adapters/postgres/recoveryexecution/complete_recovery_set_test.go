package recoveryexecution

import (
	"strings"
	"testing"
)

func TestCompleteRecoverySetUsesEvidenceRoutine(t *testing.T) {
	if !strings.Contains(completeRecoverySetSQL, "platformops_complete_recovery_execution_v2") {
		t.Fatal("recovery completion routine is not bound")
	}
}
