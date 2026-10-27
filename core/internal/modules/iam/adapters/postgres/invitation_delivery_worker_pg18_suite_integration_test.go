//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestIAMInvitationDeliveryWorkerAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, "invitation_delivery_"+origin)
			applyIAMFoundation(t, ctx, database, origin)
			applyIAMInvitationPlatform(t, ctx, database, origin)
			applyIAMLifecycleGuards(t, ctx, database, origin)
			applyIAMRBAC(t, ctx, database, origin)
			applyIAMPermissionInvitation(t, ctx, database, origin)
			prepareIAMAtomicRuntimeRole(t, ctx, database)
			seedIAMFoundationBase(t, ctx, database)
			for _, tenant := range iamFoundationTenants {
				seedIAMRBACCatalog(t, ctx, database, tenant.schema)
				seedIAMRBACProtectedPolicy(t, ctx, database, tenant)
				seedIAMPermissionInvitation(t, ctx, database, tenant)
			}
			applyIAMProfileRevisionMigrations(t, ctx, database, origin)
			applyIAMInvitationHTTPRuntime(t, ctx, database, origin)
			applyIAMInvitationDeliveryWorker(t, ctx, database, origin)
			fixtures := []iamInvitationRegistrationFixture{
				newIAMInvitationRegistrationFixture(iamFoundationTenants[0], "9a"),
				newIAMInvitationRegistrationFixture(iamFoundationTenants[1], "9b"),
			}
			for index, tenant := range iamFoundationTenants {
				createIAMInvitationRegistrationV2(t, ctx, database, tenant, fixtures[index])
			}
			assertIAMInvitationDeliveryAtomicity(
				t, ctx, database, iamFoundationTenants[0], fixtures[0],
			)
			assertIAMInvitationDeliveryIsolation(
				t, ctx, database, iamFoundationTenants[1], fixtures[1],
			)
			assertIAMInvitationDeliveryClaimState(
				t, ctx, database, iamFoundationTenants[1], fixtures[1],
			)
		})
	}
}
