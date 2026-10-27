package organizationpostgres

import (
	"strings"
	"testing"
)

func TestUpdateOrganizationDelegatesUpdateRoutine(t *testing.T) {
	if !strings.Contains(updateOrganizationSQL, "FROM iam_identity_organization_update_v3(") ||
		!strings.Contains(updateOrganizationSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("update organization SQL bypasses the production tenant routine")
	}
}
