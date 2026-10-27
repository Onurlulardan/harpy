//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

type analyticsCatalogHeadState struct {
	projectionID     string
	packageVersion   string
	resourceID       string
	publishedID      string
	componentState   string
	operationalState string
}

func readAnalyticsCatalogEndpointHead(
	t *testing.T, ctx context.Context, database *sql.DB, componentID string,
) analyticsCatalogHeadState {
	t.Helper()
	var result analyticsCatalogHeadState
	err := database.QueryRowContext(ctx, `SELECT head.projection_receipt_id::text,
projection.package_version,projection.operational_resource_id::text,
projection.published_version_id::text,component.lifecycle_state,root.lifecycle_state
FROM t_analytics_a.analytics_catalog_component_heads AS head
JOIN t_analytics_a.analytics_catalog_version_projections AS projection
  ON projection.receipt_id=head.projection_receipt_id
JOIN t_analytics_a.analytics_catalog_components AS component
  ON component.component_id=head.component_id
JOIN t_analytics_a.analytics_saved_queries AS root
  ON root.id=projection.operational_resource_id
WHERE head.component_id=$1`, componentID).Scan(
		&result.projectionID, &result.packageVersion, &result.resourceID,
		&result.publishedID, &result.componentState, &result.operationalState,
	)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func assertAnalyticsCatalogAttemptSuperseded(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	componentID string,
	projectionID string,
	newAttemptID string,
) {
	t.Helper()
	var candidateAttempt string
	var unstamped bool
	var attempts int
	err := database.QueryRowContext(ctx, `SELECT candidate.attempt_receipt_id::text,
candidate.staged_operation_id IS NULL,
(SELECT count(*) FROM t_analytics_a.analytics_catalog_projection_attempts
 WHERE projection_receipt_id=$2)
FROM t_analytics_a.analytics_catalog_component_candidates AS candidate
WHERE candidate.component_id=$1`, componentID, projectionID).Scan(
		&candidateAttempt, &unstamped, &attempts,
	)
	if err != nil {
		t.Fatal(err)
	}
	if candidateAttempt != newAttemptID || !unstamped || attempts != 2 {
		t.Fatalf("Analytics candidate supersession = %s/%t/%d",
			candidateAttempt, unstamped, attempts)
	}
}

func assertAnalyticsCatalogHeadSwitch(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	componentID string,
	first analyticsCatalogUpgradeProjection,
	second analyticsCatalogUpgradeProjection,
) {
	t.Helper()
	head := readAnalyticsCatalogEndpointHead(t, ctx, database, componentID)
	var oldState string
	var activations, liveActivations int
	if err := database.QueryRowContext(ctx, `SELECT lifecycle_state
FROM t_analytics_a.analytics_saved_queries WHERE id=$1`,
		first.resourceID).Scan(&oldState); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRowContext(ctx, `SELECT count(*),
count(*) FILTER (WHERE superseded_at IS NULL)
FROM t_analytics_a.analytics_catalog_projection_activations
WHERE component_id=$1`, componentID).Scan(
		&activations, &liveActivations,
	); err != nil {
		t.Fatal(err)
	}
	if first.resourceID == second.resourceID ||
		first.publishedID == second.publishedID ||
		head.projectionID != second.projectionID ||
		head.resourceID != second.resourceID ||
		head.publishedID != second.publishedID ||
		head.packageVersion != "2.0.0" ||
		head.componentState != "ACTIVE" ||
		head.operationalState != "active" ||
		oldState != "retired" || activations != 2 || liveActivations != 1 {
		t.Fatalf("Analytics v1-v2 head switch = %+v old=%s history=%d/%d",
			head, oldState, activations, liveActivations)
	}
}

func assertAnalyticsCatalogDisabledUpgrade(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	fixture analyticsCatalogProjectionFixture,
	projection analyticsCatalogUpgradeProjection,
) {
	t.Helper()
	head := readAnalyticsCatalogEndpointHead(t, ctx, database, fixture.componentID)
	if head.projectionID != projection.projectionID ||
		head.resourceID != projection.resourceID ||
		head.packageVersion != "3.0.0" ||
		head.componentState != "DISABLED" ||
		head.operationalState != "retired" {
		t.Fatalf("disabled Analytics upgrade head = %+v", head)
	}
}

func assertAnalyticsDuplicateDashboardWidgets(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	projection analyticsCatalogUpgradeProjection,
) {
	t.Helper()
	var count int
	if err := database.QueryRowContext(ctx, `SELECT count(*)
FROM t_analytics_a.analytics_dashboard_widgets
WHERE dashboard_version_id=$1`, projection.publishedID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("duplicate-source Analytics dashboard widgets = %d, want 2", count)
	}
}
