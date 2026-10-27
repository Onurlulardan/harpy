//go:build integration

package postgres

import (
	"context"
	"testing"
)

func TestIAMPlatformGuardsAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx := context.Background()
			database := startIAMPlatformGuardPG18(t, ctx, origin)
			prepareIAMPlatformGuardSchema(t, ctx, database, origin)
			assertIAMPlatformGuardCatalog(t, ctx, database)
			assertIAMThrottlePolicyMatrix(t, ctx, database)
			selector := assertIAMRefreshLocatorGuards(t, ctx, database)
			assertIAMTargetResolvers(t, ctx, database, selector)
			assertIAMThrottleGuards(t, ctx, database, selector)
			assertIAMPlatformGuardBypasses(t, ctx, database)
			assertIAMPlatformGuardRaces(t, ctx, database)
			assertIAMThrottleClockRollback(t, ctx, database)
		})
	}
}
