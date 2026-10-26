//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"
)

func TestAnalyticsDashboardExportPlatformUnknownAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startAnalyticsPG18(t, ctx, origin)
			prepareAnalyticsPrerequisites(t, ctx, database, origin)
			applyAnalyticsDocuments(t, ctx, database, "",
				analyticsDocuments(t, origin, "platform"))
			reconciled, remaining := reconcileAnalyticsAuthorities(t, ctx, database, true)
			if reconciled != 2 || remaining != 0 {
				t.Fatalf("Analytics authority reconcile = %d/%d", reconciled, remaining)
			}
			for _, schema := range []string{"t_analytics_a", "t_analytics_b"} {
				applyAnalyticsDocuments(t, ctx, database, schema,
					analyticsDocuments(t, origin, "tenant"))
			}
			assertAnalyticsCatalog(t, ctx, database)
			assertAnalyticsTenantIsolation(t, ctx, database)
			assertAnalyticsQueryPublishing(t, ctx, database)
			assertAnalyticsDashboardPublishing(t, ctx, database)
			prepareAnalyticsDashboardExportPlatformLifecycle(t, ctx, database, origin)
			connection := analyticsRuntimeConnection(t, ctx, database)
			occurredAt := time.Now().UTC().Add(-2 * time.Second).Truncate(time.Microsecond)
			submitAnalyticsDashboardExportFixture(t, ctx, connection,
				analyticsUnknownOperationV2, analyticsUnknownExportV2,
				analyticsUnknownJobV2, analyticsUnknownExpiryJobV2,
				"analytics-unknown-platform-v1", 0xd1, occurredAt,
				occurredAt.Add(24*time.Hour))
			assertAnalyticsDashboardExportUnknownLifecycle(t, ctx, database, connection)
		})
	}
}
