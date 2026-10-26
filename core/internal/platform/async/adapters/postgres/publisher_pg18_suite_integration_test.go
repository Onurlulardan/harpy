//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestOutboxPublisherStoreAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			fixture := newPublisherPGFixture(t, ctx, origin)
			assertPublisherOrderingAndLifecycle(t, ctx, fixture)
			assertPublisherClaimRace(t, ctx, fixture)
			assertPublisherLeaseRecovery(t, ctx, fixture)
			assertPublisherStaleGeneration(t, ctx, fixture)
			assertPublisherStaleParking(t, ctx, fixture)
			assertPublisherStaleParkingRace(t, ctx, fixture)
			assertPublisherCatalogFence(t, ctx, fixture)
		})
	}
}
