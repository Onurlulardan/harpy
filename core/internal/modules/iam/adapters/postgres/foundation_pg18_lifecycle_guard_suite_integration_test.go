//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestIAMLifecycleGuardsAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, "guards_"+origin)
			applyIAMFoundation(t, ctx, database, origin)
			seedIAMFoundationBase(t, ctx, database)
			assertIAMHumanAccessConstraints(t, ctx, database)
			assertIAMInvitationTokenConstraints(t, ctx, database)
			assertIAMSessionRefreshConstraints(t, ctx, database)
			assertIAMAuthenticatorChallengeConstraints(t, ctx, database)
			applyIAMLifecycleGuards(t, ctx, database, origin)
			for _, tenant := range iamFoundationTenants {
				t.Run(tenant.schema, func(t *testing.T) {
					assertIAMLifecycleGuardCatalog(t, ctx, database, tenant.schema)
					assertIAMLifecycleGuardMonotonicity(t, ctx, database, tenant.schema)
					assertIAMLifecycleGuardTerminality(t, ctx, database, tenant.schema)
				})
			}
		})
	}
}
