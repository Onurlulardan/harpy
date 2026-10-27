package restoreexecution

import (
	"strings"
	"testing"
)

func TestMarkSwitchedBindsVerificationDigests(t *testing.T) {
	if !strings.Contains(markSwitchedSQL, "platformops_mark_restore_switched_v2") ||
		!strings.Contains(markSwitchedSQL, "$11") {
		t.Fatal("switch does not bind all verification digests")
	}
}
