package recoveryevidence

import (
	"strings"
	"testing"
)

func TestReadOwnerAnchorsUsesOwnerPublicRoutine(t *testing.T) {
	if !strings.Contains(
		readOwnerAnchorsSQL,
		"iam_recovery_owner_anchor_evidence_v1",
	) {
		t.Fatal("owner anchor evidence bypasses the IAM routine")
	}
}
