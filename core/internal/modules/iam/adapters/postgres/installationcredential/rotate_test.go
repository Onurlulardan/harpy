package installationcredentialpostgres

import (
	"strings"
	"testing"
)

func TestRotateUsesOwnerRoutineWithoutPlaintext(t *testing.T) {
	if !strings.Contains(rotateSQL, "iam_installation_credential_supersede_v1") ||
		strings.Contains(strings.ToLower(rotateSQL), "password") {
		t.Fatal("installation credential supersede SQL contract drifted")
	}
}
