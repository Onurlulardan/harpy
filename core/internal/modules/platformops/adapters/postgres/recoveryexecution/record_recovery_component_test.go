package recoveryexecution

import (
	"strings"
	"testing"
)

func TestRecordRecoveryComponentUsesClaimBoundRoutine(t *testing.T) {
	if !strings.Contains(recordRecoveryComponentSQL, "platformops_record_recovery_component_v3") {
		t.Fatal("claim-bound component routine is not bound")
	}
}
