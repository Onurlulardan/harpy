package recoveryexecution

import (
	"strings"
	"testing"
)

func TestRejectFailureUsesLeaseBoundRoutine(t *testing.T) {
	if !strings.Contains(
		failRecoveryExecutionSQL,
		"platformops_fail_recovery_execution_v1",
	) {
		t.Fatal("recovery failure is not lease-bound")
	}
}
