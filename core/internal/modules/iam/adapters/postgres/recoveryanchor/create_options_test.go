package recoveryanchorpostgres

import (
	"strings"
	"testing"
)

func TestCreateOptionsUsesOwnerRoutine(t *testing.T) {
	if !strings.Contains(createOptionsSQL, "iam_recovery_anchor_options_v1") {
		t.Fatal("recovery anchor options SQL contract drifted")
	}
}
