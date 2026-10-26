//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"
)

func TestAnalyticsMigrationsAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startAnalyticsPG18(t, ctx, origin)
			prepareAnalyticsPrerequisites(t, ctx, database, origin)
			applyAnalyticsDocuments(t, ctx, database, "",
				analyticsDocuments(t, origin, "platform"))
			reconciled, remaining := reconcileAnalyticsAuthorities(
				t, ctx, database, true,
			)
			if reconciled != 2 || remaining != 0 {
				t.Fatalf("Analytics authority reconcile = %d/%d, want 2/0",
					reconciled, remaining)
			}
			for _, schema := range []string{"t_analytics_a", "t_analytics_b"} {
				applyAnalyticsDocuments(t, ctx, database, schema,
					analyticsDocuments(t, origin, "tenant"))
			}
			assertAnalyticsCatalog(t, ctx, database)
			assertAnalyticsTenantIsolation(t, ctx, database)
			assertAnalyticsQueryPublishing(t, ctx, database)
			assertAnalyticsDashboardPublishing(t, ctx, database)
			assertAnalyticsDashboardExport(t, ctx, database)
			assertAnalyticsDashboardDefaultRead(t, ctx, database)
			assertAnalyticsPublicLimiter(t, ctx, database)
			assertAnalyticsLegacyRetirement(t, ctx, database)
			assertAnalyticsPublicationRepair(t, ctx, database)
			assertAnalyticsRetention(t, ctx, database)
			assertAnalyticsV1Adoption(t, ctx, database)
		})
	}
}
