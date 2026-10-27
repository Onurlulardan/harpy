//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestIAMProfileCommandsAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, "profile_command_"+origin)
			applyIAMFoundation(t, ctx, database, origin)
			applyIAMLifecycleGuards(t, ctx, database, origin)
			seedIAMFoundationBase(t, ctx, database)
			applyIAMProfileRevisionMigrations(t, ctx, database, origin)
			applyIAMRBAC(t, ctx, database, origin)
			applyIAMPermissionInvitation(t, ctx, database, origin)
			applyIAMProfileCommands(t, ctx, database, origin)
			prepareIAMAtomicRuntimeRole(t, ctx, database)
			for _, tenant := range iamFoundationTenants {
				seedIAMRBACCatalog(t, ctx, database, tenant.schema)
				seedIAMRBACProtectedPolicy(t, ctx, database, tenant)
				seedIAMPermissionInvitation(t, ctx, database, tenant)
			}
			fixtures := []iamProfileCommandFixture{
				newIAMProfileCommandFixture(t, 96, "Tenant A"),
				newIAMProfileCommandFixture(t, 128, "Tenant B"),
			}
			for index, tenant := range iamFoundationTenants {
				marker := byte(0xa1 + index)
				assertIAMProfilePublishRuntime(
					t, ctx, database, tenant, fixtures[index], marker,
				)
				assertIAMProfileBackfillRuntime(
					t, ctx, database, tenant, fixtures[index], marker,
				)
			}
			assertIAMProfileRetireRuntime(
				t, ctx, database, iamFoundationTenants[0], fixtures[0], 0xa1,
			)
			assertIAMProfileRetireTenantIsolation(
				t, ctx, database, iamFoundationTenants[1],
			)
			assertIAMProfileDirectDMLDenied(
				t, ctx, database, iamFoundationTenants[1],
			)
		})
	}
}
