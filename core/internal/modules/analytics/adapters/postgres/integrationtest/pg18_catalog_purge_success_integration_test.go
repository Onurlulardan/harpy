//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func assertAnalyticsCatalogPurgeSuccess(
	t *testing.T, ctx context.Context, database *sql.DB,
	connection *sql.Conn, fixture analyticsCatalogPurgeFixture, epoch time.Time,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `
INSERT INTO t_analytics_a.analytics_dashboard_acls(
 dashboard_id,principal_kind,principal_id,access_level,row_revision,granted_by,granted_at
) VALUES($1,'role',$2,'read',1,$3,$4)`,
		fixture.dashboardProjection.resourceID, analyticsUpgradeUUID(2200),
		analyticsCatalogActorID, epoch.Add(20*time.Microsecond),
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = database.ExecContext(ctx, `
INSERT INTO t_analytics_a.analytics_dashboard_preferences(
 principal_id,surface,dashboard_id,row_revision,updated_by,updated_at
) VALUES($2,'home',$1,1,$3,$4)`,
		fixture.dashboardProjection.resourceID, analyticsUpgradeUUID(2200),
		analyticsCatalogActorID, epoch.Add(20*time.Microsecond),
	)
	if err != nil {
		t.Fatal(err)
	}
	planOperation := analyticsUpgradeUUID(2210)
	requireAnalyticsCatalogCommand(
		t, ctx, connection, fixture.endpoint, analyticsUpgradeUUID(2211),
		planOperation, "retire", 1,
		epoch.Add(21*time.Microsecond),
	)
	requireAnalyticsCatalogCommand(
		t, ctx, connection, fixture.dashboard, analyticsUpgradeUUID(2212),
		planOperation, "retire", 1,
		epoch.Add(22*time.Microsecond),
	)
	requireAnalyticsCatalogZeroBlockers(
		t, ctx, connection, fixture.endpoint, analyticsUpgradeUUID(2215),
		planOperation, epoch.Add(23*time.Microsecond),
	)
	requireAnalyticsCatalogZeroBlockers(
		t, ctx, connection, fixture.dashboard, analyticsUpgradeUUID(2216),
		planOperation, epoch.Add(24*time.Microsecond),
	)
	holdID := analyticsUpgradeUUID(2217)
	if err = placeAnalyticsLegalHold(
		ctx, connection, holdID, epoch.Add(25*time.Microsecond),
	); err != nil {
		t.Fatal(err)
	}
	endpointPurgeReceipt := analyticsUpgradeUUID(2218)
	dashboardPurgeReceipt := analyticsUpgradeUUID(2219)
	err = callAnalyticsCatalogCommand(
		ctx, connection, fixture.endpoint, endpointPurgeReceipt,
		planOperation, "purge", 1, epoch.Add(26*time.Microsecond),
	)
	requireAnalyticsPGCode(t, err, "55000", "purge legal-hold recheck")
	assertAnalyticsCatalogPurgeUntouched(
		t, ctx, database, fixture.endpoint, fixture.endpointProjection,
	)
	if err = releaseAnalyticsLegalHold(
		ctx, connection, holdID, epoch.Add(27*time.Microsecond),
	); err != nil {
		t.Fatal(err)
	}
	err = applyAnalyticsCatalogPurgePlan(
		ctx, connection, fixture, endpointPurgeReceipt,
		dashboardPurgeReceipt, planOperation, epoch.Add(28*time.Microsecond),
	)
	if err != nil {
		t.Fatalf("Analytics Catalog atomic purge plan: %v", err)
	}
	requireAnalyticsCatalogCommand(
		t, ctx, connection, fixture.endpoint, endpointPurgeReceipt,
		planOperation, "purge", 1, epoch.Add(30*time.Microsecond),
	)
	requireAnalyticsCatalogCommand(
		t, ctx, connection, fixture.dashboard, dashboardPurgeReceipt,
		planOperation, "purge", 1, epoch.Add(31*time.Microsecond),
	)
	assertAnalyticsCatalogPurgedGraph(
		t, ctx, database, fixture.endpoint, fixture.endpointProjection,
	)
	assertAnalyticsCatalogPurgedGraph(
		t, ctx, database, fixture.dashboard, fixture.dashboardProjection,
	)
}

func assertAnalyticsCatalogPurgeUntouched(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture analyticsCatalogProjectionFixture,
	projection analyticsCatalogUpgradeProjection,
) {
	t.Helper()
	var roots, heads int
	var state string
	query := `SELECT
(SELECT count(*) FROM t_analytics_a.analytics_dashboards WHERE id=$1),
(SELECT count(*) FROM t_analytics_a.analytics_catalog_component_heads
 WHERE component_id=$2),lifecycle_state
FROM t_analytics_a.analytics_catalog_components WHERE component_id=$2`
	if fixture.kind == "endpoint" {
		query = `SELECT
(SELECT count(*) FROM t_analytics_a.analytics_saved_queries WHERE id=$1),
(SELECT count(*) FROM t_analytics_a.analytics_catalog_component_heads
 WHERE component_id=$2),lifecycle_state
FROM t_analytics_a.analytics_catalog_components WHERE component_id=$2`
	}
	if err := database.QueryRowContext(
		ctx, query, projection.resourceID, fixture.componentID,
	).Scan(&roots, &heads, &state); err != nil {
		t.Fatal(err)
	}
	if roots != 1 || heads != 1 || state != "RETIRED" {
		t.Fatalf("Analytics failed purge changed graph = %d/%d/%s",
			roots, heads, state)
	}
}

func assertAnalyticsCatalogPurgedGraph(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture analyticsCatalogProjectionFixture,
	projection analyticsCatalogUpgradeProjection,
) {
	t.Helper()
	var roots, versions, children, heads, liveActivations, receipts int
	var state string
	query := `SELECT
(SELECT count(*) FROM t_analytics_a.analytics_dashboards WHERE id=$1),
(SELECT count(*) FROM t_analytics_a.analytics_dashboard_versions
 WHERE dashboard_id=$1),
(SELECT count(*) FROM t_analytics_a.analytics_dashboard_widgets
 WHERE dashboard_version_id=$3)
 +(SELECT count(*) FROM t_analytics_a.analytics_dashboard_acls WHERE dashboard_id=$1)
 +(SELECT count(*) FROM t_analytics_a.analytics_dashboard_preferences WHERE dashboard_id=$1),
(SELECT count(*) FROM t_analytics_a.analytics_catalog_component_heads
 WHERE component_id=$2),
(SELECT count(*) FROM t_analytics_a.analytics_catalog_projection_activations
 WHERE component_id=$2 AND superseded_at IS NULL),
(SELECT count(*) FROM t_analytics_a.analytics_catalog_component_receipts
 WHERE component_id=$2 AND action='purge'),component.lifecycle_state
FROM t_analytics_a.analytics_catalog_components AS component
WHERE component.component_id=$2`
	if fixture.kind == "endpoint" {
		query = `SELECT
(SELECT count(*) FROM t_analytics_a.analytics_saved_queries WHERE id=$1),
(SELECT count(*) FROM t_analytics_a.analytics_saved_query_versions WHERE query_id=$1),
(SELECT count(*) FROM t_analytics_a.analytics_query_dependencies
 WHERE query_version_id=$3),
(SELECT count(*) FROM t_analytics_a.analytics_catalog_component_heads
 WHERE component_id=$2),
(SELECT count(*) FROM t_analytics_a.analytics_catalog_projection_activations
 WHERE component_id=$2 AND superseded_at IS NULL),
(SELECT count(*) FROM t_analytics_a.analytics_catalog_component_receipts
 WHERE component_id=$2 AND action='purge'),component.lifecycle_state
FROM t_analytics_a.analytics_catalog_components AS component
WHERE component.component_id=$2`
	}
	err := database.QueryRowContext(
		ctx, query, projection.resourceID, fixture.componentID,
		projection.publishedID,
	).Scan(&roots, &versions, &children, &heads, &liveActivations, &receipts, &state)
	if err != nil {
		t.Fatal(err)
	}
	if roots+versions+children+heads+liveActivations != 0 ||
		receipts != 1 || state != "PURGED" {
		t.Fatalf("Analytics purged graph = %d/%d/%d/%d/%d receipts=%d state=%s",
			roots, versions, children, heads, liveActivations, receipts, state)
	}
}
