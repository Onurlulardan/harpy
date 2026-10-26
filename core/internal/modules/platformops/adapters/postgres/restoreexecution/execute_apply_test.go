package restoreexecution

import (
	"strings"
	"testing"
)

func TestExecuteApplyBindsDrainAcknowledgement(t *testing.T) {
	if !strings.Contains(beginApplySQL, "platformops_start_restore_apply_v2") ||
		!strings.Contains(beginApplySQL, "$6") {
		t.Fatal("apply does not bind the drain acknowledgement")
	}
}
