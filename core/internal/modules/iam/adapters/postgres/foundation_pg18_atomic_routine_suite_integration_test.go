//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestIAMAtomicRoutinesAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, "atomic_"+origin)
			applyIAMFoundation(t, ctx, database, origin)
			seedIAMFoundationBase(t, ctx, database)
			assertIAMHumanAccessConstraints(t, ctx, database)
			assertIAMInvitationTokenConstraints(t, ctx, database)
			assertIAMSessionRefreshConstraints(t, ctx, database)
			assertIAMAuthenticatorChallengeConstraints(t, ctx, database)
			seedIAMAtomicRoutineFixtures(t, ctx, database)
			prepareIAMAtomicRuntimeRole(t, ctx, database)
			applyIAMLifecycleGuards(t, ctx, database, origin)
			applyIAMAtomicRoutines(t, ctx, database, origin)
			assertIAMAtomicRoutineACL(t, ctx, database)
			assertIAMAtomicRoutineScopeSecurity(t, ctx, database)
			for _, tenant := range iamFoundationTenants {
				assertIAMAtomicRoutineCommands(t, ctx, database, tenant.schema)
			}
			assertIAMRefreshRotationRace(t, ctx, database)
		})
	}
}
