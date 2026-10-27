//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

func TestIAMProfileRevisionAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			database := startIAMFoundationPG18(t, ctx, "profile_"+origin)
			applyIAMFoundation(t, ctx, database, origin)
			seedIAMFoundationBase(t, ctx, database)
			prepareIAMAtomicRuntimeRole(t, ctx, database)
			seedIAMProfileRevisionFixtures(t, ctx, database)
			applyIAMProfileRevisionMigrations(t, ctx, database, origin)

			t.Run("backfill_and_hashes", func(t *testing.T) {
				assertIAMProfileRevisionBackfill(t, ctx, database)
			})
			t.Run("immutable_guards", func(t *testing.T) {
				assertIAMProfileRevisionGuards(t, ctx, database)
			})
			t.Run("publish_next_schema", func(t *testing.T) {
				publishIAMProfileSchemaV2(t, ctx, database)
			})
			t.Run("historical_and_active_reads", func(t *testing.T) {
				assertIAMProfileRevisionReads(t, ctx, database)
			})
			t.Run("authority", func(t *testing.T) {
				assertIAMProfileRevisionAuthority(t, ctx, database)
			})
		})
	}
}
