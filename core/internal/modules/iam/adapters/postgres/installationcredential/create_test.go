package installationcredentialpostgres

import (
	"strings"
	"testing"
)

func TestCreateUsesOwnerRoutineWithoutPlaintext(t *testing.T) {
	if !strings.Contains(createSQL, "iam_installation_credential_initialize_v2") ||
		strings.Contains(strings.ToLower(createSQL), "password") {
		t.Fatal("installation credential initialize SQL contract drifted")
	}
}
