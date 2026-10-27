//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestIAMPermissionInvitationRuntimeAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, "permission_invitation_"+origin)
			applyIAMFoundation(t, ctx, database, origin)
			applyIAMInvitationPlatform(t, ctx, database, origin)
			applyIAMLifecycleGuards(t, ctx, database, origin)
			applyIAMRBAC(t, ctx, database, origin)
			applyIAMPermissionInvitation(t, ctx, database, origin)
			prepareIAMAtomicRuntimeRole(t, ctx, database)
			seedIAMFoundationBase(t, ctx, database)
			for index, tenant := range iamFoundationTenants {
				seedIAMRBACCatalog(t, ctx, database, tenant.schema)
				seedIAMRBACProtectedPolicy(t, ctx, database, tenant)
				seedIAMPermissionInvitation(t, ctx, database, tenant)
				marker := "8a"
				if index == 1 {
					marker = "8b"
				}
				assertIAMPermissionRequestRuntime(t, ctx, database, tenant.schema)
				assertIAMInvitationRuntime(t, ctx, database, tenant, marker)
			}
			assertIAMPermissionInvitationTenantIsolation(t, ctx, database)
			applyIAMProfileRevisionMigrations(t, ctx, database, origin)
			applyIAMInvitationHTTPRuntime(t, ctx, database, origin)
			for _, tenant := range iamFoundationTenants {
				seedIAMPermissionRequestCatalog(t, ctx, database, tenant.schema)
				assertIAMPermissionRequestCatalogRuntime(
					t, ctx, database, tenant.schema,
				)
			}
			for index, tenant := range iamFoundationTenants {
				marker := "8a"
				if index == 1 {
					marker = "8b"
				}
				assertIAMInvitationRegistrationV2(t, ctx, database, tenant, marker)
			}
		})
	}
}
