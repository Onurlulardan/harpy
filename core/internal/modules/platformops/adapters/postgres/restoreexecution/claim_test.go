package restoreexecution

import (
	"strings"
	"testing"
)

func TestClaimUsesProfileFencedRestoreRoutine(t *testing.T) {
	for _, required := range []string{
		"platformops_claim_restore_execution_v5", "$1", "$4",
	} {
		if !strings.Contains(claimSQL, required) {
			t.Fatalf("claim SQL missing %q", required)
		}
	}
}
