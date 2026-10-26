package securityepoch

import (
	"strings"
	"testing"
)

func TestAssertionUsesOnlyNarrowSecurityEpochRoutine(t *testing.T) {
	if !strings.Contains(assertSQL, "iam_security_epoch_assert_v1(") ||
		strings.Contains(strings.ToUpper(assertSQL), "SELECT SECURITY_EPOCH") ||
		strings.Contains(strings.ToUpper(assertSQL), "UPDATE ") {
		t.Fatal("security epoch adapter bypasses the narrow assertion routine")
	}
}
