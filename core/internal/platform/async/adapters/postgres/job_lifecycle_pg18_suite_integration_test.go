//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestJobLifecycleStoreAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			fixture := newJobLifecyclePGFixture(t, ctx, origin)
			t.Run("claim heartbeat and cancellation", func(t *testing.T) {
				assertJobClaimHeartbeatAndCancellation(t, ctx, fixture)
			})
			t.Run("bound database atomicity", func(t *testing.T) {
				assertBoundDatabaseJobAtomicity(t, ctx, fixture)
			})
			t.Run("cross-kind child atomicity", func(t *testing.T) {
				assertCrossKindChildJobAtomicity(t, ctx, fixture)
			})
			t.Run("external receipt and unknown", func(t *testing.T) {
				assertExternalJobOutcomes(t, ctx, fixture)
			})
			t.Run("expired lease recovery race", func(t *testing.T) {
				assertExpiredJobRecoveryRace(t, ctx, fixture)
			})
			t.Run("catalog claim fence", func(t *testing.T) {
				assertJobCatalogClaimFence(t, ctx, fixture)
			})
			t.Run("stale generation park and restart", func(t *testing.T) {
				assertStaleJobParkAndRestart(t, ctx, fixture)
			})
			t.Run("stale external unknown race", func(t *testing.T) {
				assertStaleExternalUnknownRace(t, ctx, fixture)
			})
			t.Run("terminal dead-letter atomicity", func(t *testing.T) {
				assertTerminalJobDeadLetterAtomicity(t, ctx, fixture)
			})
			t.Run("terminal dead-letter concurrency", func(t *testing.T) {
				assertTerminalJobDeadLetterConcurrency(t, ctx, fixture)
			})
			t.Run("terminal root collision fencing", func(t *testing.T) {
				assertTerminalJobRootCollisionFailsClosed(t, ctx, fixture)
			})
			t.Run("event roots remain per consumer", func(t *testing.T) {
				assertEventRootsRemainPerConsumer(t, ctx, fixture)
			})
		})
	}
}
