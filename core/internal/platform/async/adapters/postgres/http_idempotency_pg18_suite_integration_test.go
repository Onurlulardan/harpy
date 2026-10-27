//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestHTTPIdempotencyStoreAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			fixture := newHTTPIdempotencyPGFixture(t, ctx, origin)
			assertHTTPIdempotencyRaceAndReplay(t, ctx, fixture)
			assertHTTPIdempotencyRollback(t, ctx, fixture)
			assertHTTPIdempotencyPrincipalScope(t, ctx, fixture)
			assertHTTPIdempotencyAsyncRetention(t, ctx, fixture)
		})
	}
}
