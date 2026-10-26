//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"
)

func TestAnalyticsCatalogLegacyBlockerReceiptUpgradePostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startAnalyticsPG18(t, ctx, "catalog-purge-rolling-"+origin)
			prepareAnalyticsPrerequisites(t, ctx, database, origin)
			applyAnalyticsDocuments(t, ctx, database, "",
				analyticsDocuments(t, origin, "platform"))
			reconciled, remaining := reconcileAnalyticsAuthorities(
				t, ctx, database, true,
			)
			if reconciled != 2 || remaining != 0 {
				t.Fatalf("Analytics authority reconcile = %d/%d", reconciled, remaining)
			}
			beforeBridge, bridge := splitAnalyticsPurgeBridgeDocuments(t, origin)
			applyAnalyticsDocuments(t, ctx, database, "t_analytics_a", beforeBridge)
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
			operationID := analyticsUpgradeUUID(2300)
			requireAnalyticsCatalogCommand(
				t, ctx, connection, fixture.endpoint, analyticsUpgradeUUID(2301),
				operationID, "retire", 1, epoch.Add(21*time.Microsecond),
			)
			requireAnalyticsCatalogCommand(
				t, ctx, connection, fixture.dashboard, analyticsUpgradeUUID(2302),
				operationID, "retire", 1, epoch.Add(22*time.Microsecond),
			)
			endpointProof := insertAnalyticsLegacyBlockerProof(
				t, ctx, database, analyticsUpgradeUUID(2303), operationID,
				fixture.endpoint.componentID, "[]", epoch.Add(23*time.Microsecond),
			)
			dashboardProof := insertAnalyticsLegacyBlockerProof(
				t, ctx, database, analyticsUpgradeUUID(2304), operationID,
				fixture.dashboard.componentID, "[]", epoch.Add(24*time.Microsecond),
			)
			blockedOperation := analyticsUpgradeUUID(2310)
			blockedProof := insertAnalyticsLegacyBlockerProof(
				t, ctx, database, analyticsUpgradeUUID(2311), blockedOperation,
				fixture.dashboard.componentID, `[{"Owner":"analytics",
"Code":"ANALYTICS_STATIC_POLICY","Digest":"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"}]`,
				epoch.Add(25*time.Microsecond),
			)
			err = callAnalyticsCatalogCommand(
				ctx, connection, fixture.endpoint, analyticsUpgradeUUID(2320),
				operationID, "purge", 1, epoch.Add(26*time.Microsecond),
			)
			requireAnalyticsPGCode(t, err, "55000", "pre-bridge v1 receipt")
			applyAnalyticsDocuments(t, ctx, database, "t_analytics_a", bridge)
			err = callAnalyticsCatalogCommand(
				ctx, connection, fixture.dashboard, analyticsUpgradeUUID(2321),
				blockedOperation, "purge", 1, epoch.Add(27*time.Microsecond),
			)
			requireAnalyticsPGCode(t, err, "55000", "nonzero v1 receipt")
			err = callAnalyticsCatalogCommand(
				ctx, connection, fixture.endpoint, analyticsUpgradeUUID(2322),
				operationID, "purge", 2, epoch.Add(28*time.Microsecond),
			)
			requireAnalyticsPGCode(t, err, "55000", "wrong legacy epoch")
			assertAnalyticsLegacyZeroEpochRollback(
				t, ctx, connection, fixture, operationID,
				analyticsUpgradeUUID(2330), epoch.Add(29*time.Microsecond),
			)
			endpointPurge := analyticsUpgradeUUID(2340)
			dashboardPurge := analyticsUpgradeUUID(2341)
			err = applyAnalyticsCatalogPurgePlan(
				ctx, connection, fixture, endpointPurge, dashboardPurge,
				operationID, epoch.Add(30*time.Microsecond),
			)
			if err != nil {
				t.Fatalf("Analytics legacy receipt purge plan: %v", err)
			}
			requireAnalyticsCatalogCommand(
				t, ctx, connection, fixture.endpoint, endpointPurge,
				operationID, "purge", 1, epoch.Add(32*time.Microsecond),
			)
			requireAnalyticsCatalogCommand(
				t, ctx, connection, fixture.dashboard, dashboardPurge,
				operationID, "purge", 1, epoch.Add(33*time.Microsecond),
			)
			requireAnalyticsLegacyBlockerProofUnchanged(
				t, ctx, database, endpointProof,
			)
			requireAnalyticsLegacyBlockerProofUnchanged(
				t, ctx, database, dashboardProof,
			)
			requireAnalyticsLegacyBlockerProofUnchanged(
				t, ctx, database, blockedProof,
			)
			assertAnalyticsCatalogPurgedGraph(
				t, ctx, database, fixture.endpoint, fixture.endpointProjection,
			)
			assertAnalyticsCatalogPurgedGraph(
				t, ctx, database, fixture.dashboard, fixture.dashboardProjection,
			)
		})
	}
}
