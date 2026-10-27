package recoveryexecution

import (
	"strings"
	"testing"
)

func TestPutComponentsUsesClaimBoundBatchRoutine(t *testing.T) {
	if !strings.Contains(
		putRecoveryComponentsSQL,
		"platformops_record_recovery_components_v1",
	) {
		t.Fatal("claim-bound component batch routine is not bound")
	}
}
