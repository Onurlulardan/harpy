package restoreexecution

import (
	"strings"
	"testing"
)

func TestUpdateLeaseUsesLeaseBoundRoutine(t *testing.T) {
	if !strings.Contains(heartbeatSQL, "platformops_heartbeat_restore_execution_v1") {
		t.Fatal("heartbeat is not lease-bound")
	}
}
