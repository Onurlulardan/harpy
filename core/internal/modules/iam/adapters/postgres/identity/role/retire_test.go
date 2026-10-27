package rolepostgres

import (
	"strings"
	"testing"
)

func TestRetireRoleDelegatesRetireRoutine(t *testing.T) {
	if !strings.Contains(retireRoleSQL, "FROM iam_identity_role_retire_v1(") ||
		!strings.Contains(retireRoleSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("retire role SQL bypasses the production tenant routine")
	}
}
