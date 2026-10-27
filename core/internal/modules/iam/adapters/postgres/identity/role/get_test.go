package rolepostgres

import (
	"strings"
	"testing"
)

func TestGetRoleDelegatesReadRoutine(t *testing.T) {
	if !strings.Contains(getRoleSQL, "FROM iam_identity_role_read_v1(") ||
		!strings.Contains(getRoleSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("get role SQL bypasses the production tenant routine")
	}
}
