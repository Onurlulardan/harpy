package organizationpostgres

import (
	"strings"
	"testing"
)

func TestDeleteOrganizationDelegatesDeleteRoutine(t *testing.T) {
	if !strings.Contains(deleteOrganizationSQL, "FROM iam_identity_organization_delete_v2(") ||
		!strings.Contains(deleteOrganizationSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("delete organization SQL bypasses the production tenant routine")
	}
}
