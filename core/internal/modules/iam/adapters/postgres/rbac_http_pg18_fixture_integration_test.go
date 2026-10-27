//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const (
	iamRBACHTTPOrganizationID = "a1000000-0000-4000-8000-000000000001"
	iamRBACHTTPMembershipID   = "a2000000-0000-4000-8000-000000000001"
)

func seedIAMRBACHTTPFixtures(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	for _, tenant := range iamFoundationTenants {
		seedIAMRBACCatalog(t, ctx, database, tenant.schema)
		seedIAMRBACRuntimeAuthority(t, ctx, database, tenant.schema)
		mustExecIAMTenant(
			t, ctx, database, tenant.schema, insertIAMOrganization,
			iamRBACHTTPOrganizationID, "RBAC Organization", "rbac-organization", nil,
		)
		mustExecIAMTenant(
			t, ctx, database, tenant.schema, insertIAMMembership,
			iamRBACHTTPMembershipID, iamRBACHTTPOrganizationID, iamUserOneID,
		)
		seedIAMRBACProtectedPolicy(t, ctx, database, tenant)
	}
}
