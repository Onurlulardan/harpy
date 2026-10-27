package recovery

import (
	"strings"
	"testing"
)

func TestReleaseHoldStatementUsesRecoveryReleaseRoutine(t *testing.T) {
	if !strings.Contains(releaseHoldSQL, "documents_recovery_release_hold_v2") {
		t.Fatal("release routine is not bound")
	}
}
