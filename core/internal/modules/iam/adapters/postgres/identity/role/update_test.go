package rolepostgres

import (
	"strings"
	"testing"
)

func TestUpdateRoleDelegatesUpdateRoutine(t *testing.T) {
	if !strings.Contains(updateRoleSQL, "FROM iam_identity_role_update_v2(") ||
		!strings.Contains(updateRoleSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("update role SQL bypasses the production tenant routine")
	}
}
