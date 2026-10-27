//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"
)

func assertAnalyticsCatalogStalePeerBlocks(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture analyticsCatalogPurgeFixture, epoch time.Time,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err = transaction.ExecContext(ctx,
		`SET LOCAL search_path=t_analytics_a,platform,pg_catalog`); err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(ctx,
		`SET SESSION AUTHORIZATION harpy_v2_runtime`); err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(ctx,
		`SET ROLE harpy_runtime`); err != nil {
		t.Fatal(err)
	}
	planOperation := analyticsUpgradeUUID(2160)
	if err = queryAnalyticsCatalogCommandTx(
		ctx, transaction, fixture.endpoint, analyticsUpgradeUUID(2161),
		planOperation, "retire", 1, epoch.Add(17*time.Microsecond),
	); err != nil {
		t.Fatalf("Analytics stale-peer endpoint retire: %v", err)
	}
	if err = queryAnalyticsCatalogCommandTx(
		ctx, transaction, fixture.dashboard, analyticsUpgradeUUID(2162),
		planOperation, "retire", 1, epoch.Add(18*time.Microsecond),
	); err != nil {
		t.Fatalf("Analytics stale-peer dashboard retire: %v", err)
	}
	if _, err = transaction.ExecContext(ctx, `RESET ROLE`); err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(ctx,
		`RESET SESSION AUTHORIZATION`); err != nil {
		t.Fatal(err)
	}
	var wrongPackage, wrongRelease bool
	err = transaction.QueryRowContext(ctx, `
SELECT analytics_catalog_internal_widget_peer_v2($1,'1.0.0',$2,$3),
       analytics_catalog_internal_widget_peer_v2($4,'1.0.0',$5,$3)`,
		analyticsUpgradeUUID(2168), analyticsUpgradeUUID(2001),
		fixture.dashboardProjection.publishedID, analyticsCatalogPackageID,
		analyticsUpgradeUUID(2169),
	).Scan(&wrongPackage, &wrongRelease)
	if err != nil || wrongPackage || wrongRelease {
		t.Fatalf("Analytics unrelated peer = %t/%t: %v",
			wrongPackage, wrongRelease, err)
	}
	result, err := transaction.ExecContext(ctx, `
UPDATE analytics_catalog_projection_activations
SET superseded_at=$1
WHERE component_id=$2 AND superseded_at IS NULL`,
		epoch.Add(19*time.Microsecond), fixture.dashboard.componentID,
	)
	if err != nil {
		t.Fatalf("Analytics stale-peer activation: %v", err)
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil || affected != 1 {
		t.Fatalf("Analytics stale-peer activation rows = %d: %v", affected, rowsErr)
	}
	if _, err = transaction.ExecContext(ctx,
		`SET SESSION AUTHORIZATION harpy_v2_runtime`); err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(ctx,
		`SET ROLE harpy_runtime`); err != nil {
		t.Fatal(err)
	}
	var blockers json.RawMessage
	err = transaction.QueryRowContext(ctx, `SELECT blockers
FROM analytics_catalog_component_blockers_v1($1,$2,$3,1,1,$4)`,
		analyticsUpgradeUUID(2163), planOperation,
		fixture.endpoint.componentID, epoch.Add(20*time.Microsecond),
	).Scan(&blockers)
	if err != nil {
		t.Fatalf("Analytics stale-peer blockers: %v", err)
	}
	requireAnalyticsPurgeBlockerCodes(
		t, blockers, []string{"ANALYTICS_DASHBOARD_WIDGET_REFERENCE"},
	)
}
