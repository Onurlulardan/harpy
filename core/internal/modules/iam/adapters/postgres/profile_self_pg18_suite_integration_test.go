//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestIAMSelfProfileAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, "self_profile_"+origin)
			applyIAMFoundation(t, ctx, database, origin)
			prepareIAMIdentityRuntimeAuthority(t, ctx, database, origin)
			applyIAMLifecycleGuards(t, ctx, database, origin)
			applyIAMRBAC(t, ctx, database, origin)
			applyIAMIdentityPermissionResolver(t, ctx, database, origin)
			seedIAMFoundationBase(t, ctx, database)
			applyIAMProfileRevisionMigrations(t, ctx, database, origin)
			applyIAMIdentityRuntime(t, ctx, database, origin)
			seedIAMIdentityRuntime(t, ctx, database)
			seedIAMSelfProfileIdentifiers(t, ctx, database)
			applyIAMSelfProfile(t, ctx, database, origin)
			assertIAMSelfProfileACL(t, ctx, database)
			assertIAMSelfProfileRuntime(t, ctx, database)
			if origin == "v1_adopted" {
				assertIAMV1SnapshotUnchanged(t, ctx, database)
			}
		})
	}
}
