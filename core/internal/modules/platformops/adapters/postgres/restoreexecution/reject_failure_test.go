package restoreexecution

import (
	"strings"
	"testing"
)

func TestRejectFailurePersistsFenceDisposition(t *testing.T) {
	if !strings.Contains(failSQL, "platformops_fail_restore_execution_v1") ||
		!strings.Contains(failSQL, "$5") {
		t.Fatal("failure does not persist fence disposition")
	}
}
