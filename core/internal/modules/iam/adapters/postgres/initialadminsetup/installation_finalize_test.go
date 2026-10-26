package initialadminsetup

import (
	"strings"
	"testing"
)

func TestInstallationFinalizeUsesCapabilitySuccessor(t *testing.T) {
	if !strings.Contains(finalizeSQL, "iam_initial_admin_setup_finalize_v2") ||
		strings.Contains(finalizeSQL, "finalize_v1") {
		t.Fatal("initial admin setup finalizer does not use capability successor")
	}
}
