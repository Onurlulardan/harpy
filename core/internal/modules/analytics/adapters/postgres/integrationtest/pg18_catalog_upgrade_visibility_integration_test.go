//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func assertAnalyticsCatalogStagedInvisible(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	connection *sql.Conn,
	fixture analyticsCatalogProjectionFixture,
	projection analyticsCatalogUpgradeProjection,
) {
	t.Helper()
	var state string
	query := `SELECT lifecycle_state FROM t_analytics_a.analytics_saved_queries
WHERE id=$1`
	if fixture.kind == "dashboard" {
		query = `SELECT lifecycle_state FROM t_analytics_a.analytics_dashboards
WHERE id=$1`
	}
	if err := database.QueryRowContext(ctx, query, projection.resourceID).Scan(&state); err != nil {
		t.Fatal(err)
	}
	if state != "staged" {
		t.Fatalf("Analytics Catalog pre-activation root = %s, want staged", state)
	}
	getErr := withAnalyticsCatalogRuntimeTx(ctx, connection, func(tx *sql.Tx) error {
		var count int
		if fixture.kind == "dashboard" {
			return tx.QueryRowContext(ctx, `SELECT count(*)
FROM analytics_dashboard_get_version_v1('t_analytics_a',$1,$2)`,
				projection.resourceID, projection.publishedID).Scan(&count)
		}
		return tx.QueryRowContext(ctx, `SELECT count(*)
FROM analytics_query_get_version_v1('t_analytics_a',$1,$2)`,
			projection.resourceID, projection.publishedID).Scan(&count)
	})
	requireAnalyticsPGCode(t, getErr, "P6102", "staged Analytics get")
	manifestKind := "query"
	if fixture.kind == "dashboard" {
		manifestKind = "dashboard"
	}
	manifestErr := withAnalyticsCatalogRuntimeTx(
		ctx, connection, func(tx *sql.Tx) error {
			var manifest []byte
			return tx.QueryRowContext(ctx, `SELECT analytics_resource_manifest_v1(
't_analytics_a',$1,$2)`, manifestKind, projection.publishedID).Scan(&manifest)
		},
	)
	requireAnalyticsPGCode(t, manifestErr, "P6102", "staged Analytics manifest")
}

func assertAnalyticsCatalogVisible(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	fixture analyticsCatalogProjectionFixture,
	projection analyticsCatalogUpgradeProjection,
) {
	t.Helper()
	var count int
	err := withAnalyticsCatalogRuntimeTx(ctx, connection, func(tx *sql.Tx) error {
		if fixture.kind == "dashboard" {
			return tx.QueryRowContext(ctx, `SELECT count(*)
FROM analytics_dashboard_get_version_v1('t_analytics_a',$1,$2)`,
				projection.resourceID, projection.publishedID).Scan(&count)
		}
		return tx.QueryRowContext(ctx, `SELECT count(*)
FROM analytics_query_get_version_v1('t_analytics_a',$1,$2)`,
			projection.resourceID, projection.publishedID).Scan(&count)
	})
	if err != nil {
		t.Fatalf("active Analytics Catalog resource is not visible: %v", err)
	}
	if count != 1 {
		t.Fatalf("active Analytics Catalog resource count = %d, want 1", count)
	}
}
