package restoreexecution

import (
	"strings"
	"testing"
)

func TestStartDrainUsesDedicatedTransition(t *testing.T) {
	if !strings.Contains(beginDrainSQL, "platformops_start_restore_drain_v1") {
		t.Fatal("drain uses an unexpected transition")
	}
}
