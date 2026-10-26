package rolepostgres

import (
	"strings"
	"testing"
)

func TestListRolesDelegatesListRoutine(t *testing.T) {
	if !strings.Contains(listRolesSQL, "FROM iam_identity_role_list_v1(") ||
		!strings.Contains(listRolesSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("list roles SQL bypasses the production tenant routine")
	}
}
