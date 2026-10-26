package restoreexecution

import (
	"strings"
	"testing"
)

func TestValidateRestoreBindsArchiveAndDatabaseEvidence(t *testing.T) {
	if !strings.Contains(beginVerifySQL, "platformops_start_restore_verify_v1") ||
		!strings.Contains(beginVerifySQL, "$5") {
		t.Fatal("verify does not bind archive and database evidence")
	}
}
