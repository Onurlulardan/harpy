//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"
)

func assertAnalyticsCatalogOperationalProjection(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	connection := analyticsRuntimeConnection(t, ctx, database)
	epoch, err := time.Parse(time.RFC3339Nano, analyticsCatalogProjectionEpoch)
	if err != nil {
		t.Fatal(err)
	}
	endpoint := analyticsEndpointCatalogFixture(t)
	firstReceipt := "019f9000-0000-7000-8000-000000000011"
	pending, err := projectAnalyticsCatalogFixture(
		ctx, connection, endpoint, firstReceipt, analyticsCatalogReleaseID,
		"analytics-catalog-endpoint", epoch,
	)
	if err != nil || pending.receiptID != firstReceipt ||
		pending.lifecycle != "AVAILABLE" || pending.readyAt.Valid {
		t.Fatalf("pending Analytics endpoint = %+v: %v", pending, err)
	}
	assertAnalyticsCatalogState(t, ctx, database, endpoint.componentID,
		"pending", "ANALYTICS_FORMS_DEPENDENCY_NOT_READY")
	seedFormsCatalogDependency(t, ctx, database, "hr-employees")
	stillPending, err := projectAnalyticsCatalogFixture(
		ctx, connection, endpoint,
		"019f9000-0000-7000-8000-000000000012", analyticsCatalogReleaseID,
		"analytics-catalog-endpoint", epoch.Add(time.Microsecond),
	)
	if err != nil || stillPending.receiptID != firstReceipt ||
		stillPending.lifecycle != "AVAILABLE" || stillPending.readyAt.Valid {
		t.Fatalf("partially resolved Analytics endpoint = %+v: %v", stillPending, err)
	}
	assertAnalyticsCatalogState(t, ctx, database, endpoint.componentID,
		"pending", "ANALYTICS_FORMS_DEPENDENCY_NOT_READY")
	seedFormsCatalogDependency(t, ctx, database, "hr-departments")
	ready, err := projectAnalyticsCatalogFixture(
		ctx, connection, endpoint,
		"019f9000-0000-7000-8000-000000000013", analyticsCatalogReleaseID,
		"analytics-catalog-endpoint", epoch.Add(2*time.Microsecond),
	)
	if err != nil || ready.receiptID != firstReceipt ||
		ready.lifecycle != "AVAILABLE" || !ready.readyAt.Valid {
		t.Fatalf("ready Analytics endpoint = %+v: %v", ready, err)
	}
	assertAnalyticsCatalogState(t, ctx, database, endpoint.componentID, "ready", "")
	assertAnalyticsCatalogQueryReadback(t, ctx, database, endpoint.componentID)
	replay, err := projectAnalyticsCatalogFixture(
		ctx, connection, endpoint,
		"019f9000-0000-7000-8000-000000000014", analyticsCatalogReleaseID,
		"analytics-catalog-endpoint", epoch.Add(3*time.Microsecond),
	)
	if err != nil || replay.receiptID != firstReceipt ||
		!replay.readyAt.Valid || !replay.readyAt.Time.Equal(ready.readyAt.Time) {
		t.Fatalf("Analytics endpoint replay = %+v: %v", replay, err)
	}
	_, err = projectAnalyticsCatalogFixture(
		ctx, connection, endpoint,
		"019f9000-0000-7000-8000-000000000015",
		"019f9000-0000-7000-8000-000000000099",
		"analytics-catalog-endpoint", epoch.Add(4*time.Microsecond),
	)
	requireAnalyticsPGCode(t, err, "23505", "Analytics Catalog conflict")
	dashboard := analyticsDashboardCatalogFixture(t)
	dashboardResult, err := projectAnalyticsCatalogFixture(
		ctx, connection, dashboard,
		"019f9000-0000-7000-8000-000000000021", analyticsCatalogReleaseID,
		"analytics-catalog-dashboard", epoch.Add(5*time.Microsecond),
	)
	if err != nil || !dashboardResult.readyAt.Valid {
		t.Fatalf("ready Analytics dashboard = %+v: %v", dashboardResult, err)
	}
	assertAnalyticsCatalogState(t, ctx, database, dashboard.componentID, "ready", "")
	assertAnalyticsCatalogDashboardReadback(t, ctx, database, dashboard.componentID)
	_, err = connection.ExecContext(ctx, `UPDATE
t_analytics_a.analytics_catalog_definitions SET name=name`)
	requireAnalyticsPGCode(t, err, "42501", "Analytics Catalog table ACL")
}

func assertAnalyticsCatalogState(
	t *testing.T, ctx context.Context, database *sql.DB,
	componentID, wantState, wantBlocker string,
) {
	t.Helper()
	var state string
	var blocker sql.NullString
	if err := database.QueryRowContext(ctx, `SELECT operational_state,blocker_code
FROM t_analytics_a.analytics_catalog_definitions WHERE component_id=$1`,
		componentID).Scan(&state, &blocker); err != nil {
		t.Fatal(err)
	}
	if state != wantState || blocker.String != wantBlocker ||
		blocker.Valid != (wantBlocker != "") {
		t.Fatalf("Analytics Catalog state = %s/%+v, want %s/%s",
			state, blocker, wantState, wantBlocker)
	}
}

func assertAnalyticsCatalogQueryReadback(
	t *testing.T, ctx context.Context, database *sql.DB, componentID string,
) {
	t.Helper()
	var queryID, versionID string
	var dependencies int
	var queryDocument, manifestSHA []byte
	if err := database.QueryRowContext(ctx, `SELECT
definition.operational_resource_id::text,definition.published_version_id::text,
(SELECT count(*) FROM t_analytics_a.analytics_query_dependencies AS dependency
 WHERE dependency.query_version_id=definition.published_version_id),
version.query_document,version.manifest_sha256
FROM t_analytics_a.analytics_catalog_definitions AS definition
JOIN t_analytics_a.analytics_saved_query_versions AS version
 ON version.id=definition.published_version_id
WHERE definition.component_id=$1`, componentID).Scan(
		&queryID, &versionID, &dependencies, &queryDocument, &manifestSHA,
	); err != nil {
		t.Fatal(err)
	}
	if queryID == "" || versionID == "" || dependencies != 2 {
		t.Fatalf("Analytics query readback = %s/%s dependencies=%d",
			queryID, versionID, dependencies)
	}
	var query struct {
		Contract string `json:"contract"`
		Sources  []any  `json:"sources"`
	}
	if err := json.Unmarshal(queryDocument, &query); err != nil ||
		query.Contract != "analytics.native-join-query.v1" || len(query.Sources) != 2 {
		t.Fatalf("Analytics native query readback = %+v: %v", query, err)
	}
	wantManifest := analyticsCatalogExpectedDependencyManifest(t)
	if !bytes.Equal(manifestSHA, wantManifest[:]) {
		t.Fatalf("Analytics dependency manifest differs")
	}
}

func assertAnalyticsCatalogDashboardReadback(
	t *testing.T, ctx context.Context, database *sql.DB, componentID string,
) {
	t.Helper()
	var dashboardID, versionID string
	var widgets int
	if err := database.QueryRowContext(ctx, `SELECT
definition.operational_resource_id::text,definition.published_version_id::text,
(SELECT count(*) FROM t_analytics_a.analytics_dashboard_widgets AS widget
 WHERE widget.dashboard_version_id=definition.published_version_id)
FROM t_analytics_a.analytics_catalog_definitions AS definition
WHERE definition.component_id=$1`, componentID).Scan(
		&dashboardID, &versionID, &widgets,
	); err != nil {
		t.Fatal(err)
	}
	if dashboardID == "" || versionID == "" || widgets != 1 {
		t.Fatalf("Analytics dashboard readback = %s/%s widgets=%d",
			dashboardID, versionID, widgets)
	}
}
