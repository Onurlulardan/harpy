//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func assertAnalyticsDashboardExportUnknownBusy(t *testing.T, ctx context.Context,
	database *sql.DB, connection *sql.Conn,
	evidence analyticsDashboardExportUnknownEvidence, observedAt time.Time,
) {
	t.Helper()
	before := analyticsDashboardExportUnknownSnapshot(t, ctx, database)
	claim, err := claimAnalyticsDashboardExportUnknownCleanup(t, ctx, connection,
		evidence, analyticsUnknownCompetingV2, observedAt)
	if err != nil || claim.state != "busy" || claim.generation.Valid ||
		claim.leaseUntil.Valid {
		t.Fatalf("competing unknown cleanup claim = %+v: %v", claim, err)
	}
	if before != analyticsDashboardExportUnknownSnapshot(t, ctx, database) {
		t.Fatal("competing unknown cleanup claim mutated evidence")
	}
}

func assertAnalyticsDashboardExportUnknownCleanupStale(t *testing.T,
	ctx context.Context, database *sql.DB, connection *sql.Conn,
	evidence analyticsDashboardExportUnknownEvidence, observedAt time.Time,
) {
	t.Helper()
	before := analyticsDashboardExportUnknownSnapshot(t, ctx, database)
	_, _, err := checkpointAnalyticsDashboardExportUnknownCleanup(t, ctx, connection,
		evidence, analyticsUnknownCompetingV2, 1, observedAt)
	requireAnalyticsPGCode(t, err, "P6047", "unknown cleanup stale owner")
	if before != analyticsDashboardExportUnknownSnapshot(t, ctx, database) {
		t.Fatal("unknown cleanup stale owner mutated evidence")
	}
}

func assertAnalyticsDashboardExportUnknownDeferredEvidence(t *testing.T,
	ctx context.Context, database *sql.DB,
	evidence analyticsDashboardExportUnknownEvidence, retryAt time.Time,
) {
	t.Helper()
	var exact bool
	err := database.QueryRowContext(ctx, `SELECT
claim.claim_generation=1 AND claim.lease_owner_id=$2 AND
 claim.lease_until=claim.updated_at AND claim.run_at=$3 AND
 claim.run_at=claim.updated_at+interval '30 seconds' AND
 claim.attempt_id=$4 AND claim.original_lease_owner_id=$5 AND
 claim.effect_id=$6 AND claim.request_sha256=encode($7::bytea,'hex') AND
 NOT EXISTS (SELECT 1 FROM platform.analytics_dashboard_export_unknown_resolutions
  WHERE job_id=$1) AND
 NOT EXISTS (SELECT 1 FROM t_analytics_a.analytics_dashboard_export_unknown_resolutions
  WHERE export_id=$8)
FROM platform.analytics_dashboard_export_unknown_cleanup_claims claim
WHERE claim.job_id=$1`, analyticsUnknownJobV2, analyticsUnknownCleanupV2,
		retryAt, evidence.claim.attemptID, analyticsUnknownLeaseV2,
		evidence.effect.effectID, evidence.effect.request,
		analyticsUnknownExportV2).Scan(&exact)
	if err != nil || !exact {
		t.Fatalf("unknown cleanup deferred evidence = %v: %v", exact, err)
	}
	assertAnalyticsDashboardExportUnknownEvidence(t, ctx, database, evidence)
}
