//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestIAMAuthSecurityBehaviorAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startIAMPlatformGuardPG18(t, ctx, "auth_security_"+origin)
			prepareIAMPlatformGuardSchema(t, ctx, database, origin)
			prepareIAMAuthTenantRuntime(t, ctx, database, origin)
			fixtures := seedIAMAuthRuntimeFixtures(t, ctx, database)
			factors := seedIAMAuthSecurityFactors(t, ctx, database)
			applyIAMLifecycleGuards(t, ctx, database.migrator, origin)
			applyIAMAtomicRoutines(t, ctx, database.migrator, origin)
			applyIAMAuthRuntime(t, ctx, database.migrator, origin)
			grantIAMAuthTenantRuntime(t, ctx, database)
			assertIAMRecoveryTenantSubstitutionRejected(
				t, ctx, database, factors,
			)
			for index, fixture := range fixtures {
				assertIAMAuthLoginAndSession(t, ctx, database, fixture)
				reauthentication := assertIAMPasswordReauthentication(
					t, ctx, database, fixture,
				)
				assertIAMPasswordReauthenticationReplayRejected(
					t, ctx, database, reauthentication,
				)
				assertIAMSessionRevocation(
					t, ctx, database, reauthentication,
				)
				assertIAMRecoveryConsumption(
					t, ctx, database, factors[index],
				)
			}
		})
	}
}
