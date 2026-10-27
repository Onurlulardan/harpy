package recoveryanchorpostgres

import (
	"strings"
	"testing"
)

func TestListUsesOwnerRoutine(t *testing.T) {
	if !strings.Contains(listSQL, "iam_recovery_anchor_list_v1") {
		t.Fatal("recovery anchor list SQL contract drifted")
	}
}
