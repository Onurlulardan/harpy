package recoverytarget

import (
	"strings"
	"testing"
)

func TestGetUsesRecoveryTargetRoutine(t *testing.T) {
	if !strings.Contains(getRecoveryTargetSQL, "tenancy_get_recovery_target_v1") {
		t.Fatal("recovery target routine is not bound")
	}
}
