//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestIAMIdentityRuntimeAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, "identity_runtime_"+origin)
			applyIAMFoundation(t, ctx, database, origin)
			prepareIAMIdentityRuntimeAuthority(t, ctx, database, origin)
			applyIAMLifecycleGuards(t, ctx, database, origin)
			applyIAMRBAC(t, ctx, database, origin)
			applyIAMIdentityPermissionResolver(t, ctx, database, origin)
			seedIAMFoundationBase(t, ctx, database)
			applyIAMProfileRevisionMigrations(t, ctx, database, origin)
			probeIAMIdentityRollback(t, ctx, database, origin)
			applyIAMIdentityRuntime(t, ctx, database, origin)
			applyIAMIdentityGrantEvidenceAuthority(t, ctx, database, origin)
			seedIAMIdentityRuntime(t, ctx, database)
			assertIAMIdentityGrantEvidenceAuthority(t, ctx, database)
			assertIAMIdentityUserReadsActiveRelations(t, ctx, database)
			assertIAMIdentityProfileReads(t, ctx, database)
			assertIAMIdentityPasswordEnrollment(t, ctx, database)
			assertIAMIdentityReplayAndCAS(t, ctx, database)
			assertIAMIdentityChangeAudit(t, ctx, database)
			assertIAMIdentityChangeRollback(t, ctx, database)
			assertIAMIdentityChangeACL(t, ctx, database)
			assertIAMIdentityTenantIsolation(t, ctx, database)
			assertIAMIdentitySystemResourceDeleteCleanup(t, ctx, database)
			assertIAMIdentityUserDeleteOwnerRollback(t, ctx, database)
			assertIAMIdentityProtectedAdmin(t, ctx, database)
			applyIAMIdentityParityRuntimeTargets(t, ctx, database, origin)
			assertIAMIdentityParityRuntime(t, ctx, database)
			if origin == "v1_adopted" {
				assertIAMV1SnapshotUnchanged(t, ctx, database)
			}
			assertIAMOrganizationOwnerParity(t, ctx, database)
		})
	}
}
