package rolepostgres

import (
	"strings"
	"testing"
)

func TestCreateRoleDelegatesCreateRoutine(t *testing.T) {
	if !strings.Contains(createRoleSQL, "FROM iam_identity_role_create_v2(") ||
		!strings.Contains(createRoleSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("create role SQL bypasses the production tenant routine")
	}
}
