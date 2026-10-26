package organizationpostgres

import (
	"strings"
	"testing"
)

func TestListOrganizationsDelegatesListRoutine(t *testing.T) {
	if !strings.Contains(listOrganizationsSQL, "FROM iam_identity_organization_list_v2(") ||
		!strings.Contains(listOrganizationsSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("list organizations SQL bypasses the production tenant routine")
	}
}
