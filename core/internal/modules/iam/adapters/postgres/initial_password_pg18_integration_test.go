//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestIAMInitialPasswordAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startIAMPlatformGuardPG18(t, ctx, "initial_password_"+origin)
			prepareIAMPlatformGuardSchema(t, ctx, database, origin)
			prepareIAMAuthTenantRuntime(t, ctx, database, origin)
			applyIAMLifecycleGuards(t, ctx, database.migrator, origin)
			applyIAMAtomicRoutines(t, ctx, database.migrator, origin)
			applyIAMInitialPassword(t, ctx, database, origin)
			grantIAMAuthTenantRuntime(t, ctx, database)
			fixtures := seedIAMInitialPasswordFixtures(t, ctx, database)
			assertIAMInitialPasswordAttemptLimit(t, ctx, database, fixtures[0])
			assertIAMInitialPasswordFinalize(t, ctx, database, fixtures[1])
		})
	}
}
