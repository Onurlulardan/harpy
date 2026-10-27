package recoveryevidence

import (
	"strings"
	"testing"
)

func TestReadHeadUsesOwnerPublicRoutine(t *testing.T) {
	if !strings.Contains(
		readHeadSQL,
		"platformauthority_recovery_head_evidence_v1",
	) {
		t.Fatal("protected authority evidence bypasses the owner routine")
	}
}
