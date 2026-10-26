//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

type analyticsCatalogPurgeFixture struct {
	endpoint            analyticsCatalogProjectionFixture
	dashboard           analyticsCatalogProjectionFixture
	endpointProjection  analyticsCatalogUpgradeProjection
	dashboardProjection analyticsCatalogUpgradeProjection
}

func TestAnalyticsCatalogPurgePostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startAnalyticsPG18(t, ctx, "catalog-purge-"+origin)
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
			seedFormsCatalogDependency(t, ctx, database, "hr-employees")
			seedFormsCatalogDependency(t, ctx, database, "hr-departments")
			connection := analyticsRuntimeConnection(t, ctx, database)
			epoch, err := time.Parse(time.RFC3339Nano, analyticsCatalogProjectionEpoch)
			if err != nil {
				t.Fatal(err)
			}
			fixture := prepareAnalyticsCatalogPurgeFixture(
				t, ctx, database, connection, epoch,
			)
			assertAnalyticsCatalogPurgeBlockers(
				t, ctx, database, fixture, epoch,
			)
			assertAnalyticsCatalogStalePeerBlocks(
				t, ctx, database, fixture, epoch,
			)
			assertAnalyticsCatalogPurgeSuccess(
				t, ctx, database, connection, fixture, epoch,
			)
		})
	}
}

func prepareAnalyticsCatalogPurgeFixture(
	t *testing.T, ctx context.Context, database *sql.DB,
	connection *sql.Conn, epoch time.Time,
) analyticsCatalogPurgeFixture {
	t.Helper()
	endpoint := analyticsEndpointCatalogFixture(t)
	dashboard := analyticsDashboardCatalogFixture(t)
	releaseID := analyticsUpgradeUUID(2001)
	catalogOperation := analyticsUpgradeUUID(2002)
	endpointProjection := requireAnalyticsCatalogUpgradeProjection(
		t, ctx, database, connection, endpoint, "1.0.0", releaseID,
		analyticsUpgradeUUID(2004), catalogOperation, epoch,
	)
	dashboardProjection := requireAnalyticsCatalogUpgradeProjection(
		t, ctx, database, connection, dashboard, "1.0.0", releaseID,
		analyticsUpgradeUUID(2005), catalogOperation, epoch,
	)
	requireAnalyticsCatalogUpgradeAction(
		t, ctx, connection, endpoint, catalogOperation, "stage", 2006,
		epoch.Add(time.Microsecond),
	)
	requireAnalyticsCatalogUpgradeAction(
		t, ctx, connection, dashboard, catalogOperation, "stage", 2007,
		epoch.Add(2*time.Microsecond),
	)
	requireAnalyticsCatalogUpgradeAction(
		t, ctx, connection, endpoint, catalogOperation, "activate", 2008,
		epoch.Add(3*time.Microsecond),
	)
	requireAnalyticsCatalogUpgradeAction(
		t, ctx, connection, dashboard, catalogOperation, "activate", 2009,
		epoch.Add(4*time.Microsecond),
	)
	return analyticsCatalogPurgeFixture{
		endpoint: endpoint, dashboard: dashboard,
		endpointProjection:  endpointProjection,
		dashboardProjection: dashboardProjection,
	}
}
