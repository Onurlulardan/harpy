package organizationpostgres

import (
	"strings"
	"testing"
)

func TestDisableOrganizationDelegatesDisableRoutine(t *testing.T) {
	if !strings.Contains(disableOrganizationSQL, "FROM iam_identity_organization_disable_v1(") ||
		!strings.Contains(disableOrganizationSQL, "(pg_catalog.current_schemas(false))[2]::name") {
		t.Fatal("disable organization SQL bypasses the production tenant routine")
	}
}
