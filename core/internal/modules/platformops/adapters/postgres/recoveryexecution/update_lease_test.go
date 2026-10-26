package recoveryexecution

import (
	"strings"
	"testing"
)

func TestUpdateLeaseUsesLeaseBoundRoutine(t *testing.T) {
	if !strings.Contains(
		heartbeatRecoveryExecutionSQL,
		"platformops_heartbeat_recovery_execution_v1",
	) {
		t.Fatal("recovery lease update is not lease-bound")
	}
}
