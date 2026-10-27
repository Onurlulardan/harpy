//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func assertAnalyticsCatalogUpgradeAcceptance(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	connection := analyticsRuntimeConnection(t, ctx, database)
	seedFormsCatalogDependency(t, ctx, database, "hr-employees")
	seedFormsCatalogDependency(t, ctx, database, "hr-departments")
	epoch, err := time.Parse(time.RFC3339Nano, analyticsCatalogProjectionEpoch)
	if err != nil {
		t.Fatal(err)
	}
	at := func(step int) time.Time {
		return epoch.Add(time.Duration(step) * time.Microsecond)
	}

	first := analyticsEndpointCatalogFixture(t)
	assertAnalyticsCatalogPlanTypeGuards(
		t, ctx, database, connection, first, at(0),
	)
	firstRelease := analyticsUpgradeUUID(1001)
	firstOperation := analyticsUpgradeUUID(1000)
	firstProjection := requireAnalyticsCatalogUpgradeProjection(
		t, ctx, database, connection, first, "1.0.0",
		firstRelease, analyticsUpgradeUUID(1002),
		firstOperation, at(0),
	)
	dashboard := duplicateAnalyticsDashboardFixture(t)
	dashboardOperation := analyticsUpgradeUUID(1010)
	dashboardProjection := requireAnalyticsCatalogUpgradeProjection(
		t, ctx, database, connection, dashboard, "1.0.0", firstRelease,
		analyticsUpgradeUUID(1011), dashboardOperation, at(0),
	)
	assertAnalyticsCatalogStagedInvisible(
		t, ctx, database, connection, first, firstProjection,
	)
	assertAnalyticsCatalogStagedInvisible(
		t, ctx, database, connection, dashboard, dashboardProjection,
	)
	requireAnalyticsCatalogUpgradeAction(
		t, ctx, connection, first, firstOperation, "stage", 1003, at(1),
	)
	requireAnalyticsCatalogUpgradeAction(
		t, ctx, connection, dashboard, dashboardOperation, "stage", 1012, at(2),
	)
	requireAnalyticsCatalogUpgradeAction(
		t, ctx, connection, first, firstOperation, "activate", 1004, at(3),
	)
	requireAnalyticsCatalogUpgradeAction(
		t, ctx, connection, dashboard, dashboardOperation, "activate", 1013, at(4),
	)
	assertAnalyticsCatalogVisible(t, ctx, connection, first, firstProjection)
	assertAnalyticsCatalogVisible(
		t, ctx, connection, dashboard, dashboardProjection,
	)
	assertAnalyticsDuplicateDashboardWidgets(
		t, ctx, database, dashboardProjection,
	)

	second := analyticsCatalogVersionFixture(t, first, "version-2", nil)
	secondRelease := analyticsUpgradeUUID(1101)
	staleOperation := analyticsUpgradeUUID(1102)
	staleProjection := requireAnalyticsCatalogUpgradeProjection(
		t, ctx, database, connection, second, "2.0.0", secondRelease,
		analyticsUpgradeUUID(1103), staleOperation, at(5),
	)
	requireAnalyticsCatalogUpgradeAction(
		t, ctx, connection, second, staleOperation, "stage", 1104, at(6),
	)
	currentOperation := analyticsUpgradeUUID(1105)
	currentProjection := requireAnalyticsCatalogUpgradeProjection(
		t, ctx, database, connection, second, "2.0.0", secondRelease,
		analyticsUpgradeUUID(1106), currentOperation, at(7),
	)
	if currentProjection.projectionID != staleProjection.projectionID {
		t.Fatal("same-version Analytics attempt created a distinct projection")
	}
	assertAnalyticsCatalogAttemptSuperseded(
		t, ctx, database, second.componentID, currentProjection.projectionID,
		currentProjection.receiptID,
	)
	err = applyAnalyticsCatalogUpgradeAction(
		ctx, connection, second, analyticsUpgradeUUID(1107),
		staleOperation, "activate", at(8),
	)
	requireAnalyticsPGCode(t, err, "55000", "stale Analytics activation")
	err = applyAnalyticsCatalogUpgradeAction(
		ctx, connection, second, analyticsUpgradeUUID(1108),
		staleOperation, "cleanup", at(9),
	)
	requireAnalyticsPGCode(t, err, "55000", "stale Analytics cleanup")
	secondDashboard := analyticsCatalogVersionFixture(
		t, dashboard, "version-2", nil,
	)
	secondDashboardOperation := analyticsUpgradeUUID(1210)
	secondDashboardProjection := requireAnalyticsCatalogUpgradeProjection(
		t, ctx, database, connection, secondDashboard, "2.0.0", secondRelease,
		analyticsUpgradeUUID(1211), secondDashboardOperation, at(10),
	)
	assertAnalyticsCatalogStagedInvisible(
		t, ctx, database, connection, secondDashboard, secondDashboardProjection,
	)
	requireAnalyticsCatalogUpgradeAction(
		t, ctx, connection, second, currentOperation, "stage", 1109, at(11),
	)
	requireAnalyticsCatalogUpgradeAction(
		t, ctx, connection, secondDashboard, secondDashboardOperation,
		"stage", 1212, at(12),
	)
	requireAnalyticsCatalogUpgradeAction(
		t, ctx, connection, second, currentOperation, "activate", 1110, at(13),
	)
	requireAnalyticsCatalogUpgradeAction(
		t, ctx, connection, secondDashboard, secondDashboardOperation,
		"activate", 1213, at(14),
	)
	assertAnalyticsCatalogVisible(t, ctx, connection, second, currentProjection)
	assertAnalyticsCatalogVisible(
		t, ctx, connection, secondDashboard, secondDashboardProjection,
	)
	assertAnalyticsCatalogHeadSwitch(
		t, ctx, database, second.componentID, firstProjection, currentProjection,
	)

	requireAnalyticsCatalogUpgradeAction(
		t, ctx, connection, second, analyticsUpgradeUUID(1300),
		"disable", 1301, at(15),
	)
	third := analyticsCatalogVersionFixture(t, second, "version-3", nil)
	thirdOperation := analyticsUpgradeUUID(1302)
	thirdProjection := requireAnalyticsCatalogUpgradeProjection(
		t, ctx, database, connection, third, "3.0.0",
		analyticsUpgradeUUID(1303), analyticsUpgradeUUID(1304),
		thirdOperation, at(16),
	)
	requireAnalyticsCatalogUpgradeAction(
		t, ctx, connection, third, thirdOperation, "stage", 1305, at(17),
	)
	requireAnalyticsCatalogUpgradeAction(
		t, ctx, connection, third, thirdOperation, "activate", 1306, at(18),
	)
	assertAnalyticsCatalogDisabledUpgrade(
		t, ctx, database, third, thirdProjection,
	)
}

func duplicateAnalyticsDashboardFixture(
	t *testing.T,
) analyticsCatalogProjectionFixture {
	t.Helper()
	return analyticsCatalogVersionFixture(
		t, analyticsDashboardCatalogFixture(t), "duplicate-source", func(definition map[string]any) {
			widgets := definition["widgets"].([]any)
			definition["widgets"] = append(widgets, map[string]any{
				"id": "employee-summary", "dataSource": "employee-source",
				"type": "metric",
			})
		},
	)
}
