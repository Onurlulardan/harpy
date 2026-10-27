package organizationpostgres

import (
	"strings"
	"testing"
)

func TestCreateOrganizationDelegatesCreateRoutine(t *testing.T) {
	if !strings.Contains(createOrganizationSQL, "FROM iam_identity_organization_create_v3(") ||
		!strings.Contains(createOrganizationSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("create organization SQL bypasses the production tenant routine")
	}
}
