package organizationpostgres

import (
	"strings"
	"testing"
)

func TestGetOrganizationDelegatesReadRoutine(t *testing.T) {
	if !strings.Contains(getOrganizationSQL, "FROM iam_identity_organization_read_v2(") ||
		!strings.Contains(getOrganizationSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("get organization SQL bypasses the production tenant routine")
	}
}
