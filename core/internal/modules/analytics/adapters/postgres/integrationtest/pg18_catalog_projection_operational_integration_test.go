//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"
)

func TestAnalyticsCatalogProjectionMaterializesOperationalResourcesPostgreSQL184(
	t *testing.T,
) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startAnalyticsPG18(t, ctx, "catalog-projection-"+origin)
			prepareAnalyticsPrerequisites(t, ctx, database, origin)
			applyAnalyticsDocuments(t, ctx, database, "",
				analyticsDocuments(t, origin, "platform"))
			reconciled, remaining := reconcileAnalyticsAuthorities(
				t, ctx, database, true,
			)
			if reconciled != 2 || remaining != 0 {
				t.Fatalf("Analytics authority reconcile = %d/%d", reconciled, remaining)
			}
			applyAnalyticsDocuments(t, ctx, database, "t_analytics_a",
				analyticsDocuments(t, origin, "tenant"))
			installFormsCatalogDependencyFixture(t, ctx, database)
			assertAnalyticsCatalogRuntimePrivileges(t, ctx, database)
			assertAnalyticsCatalogUpgradeAcceptance(t, ctx, database)
			assertAnalyticsCatalogLegacyCompatibility(t, ctx, database)
		})
	}
}
