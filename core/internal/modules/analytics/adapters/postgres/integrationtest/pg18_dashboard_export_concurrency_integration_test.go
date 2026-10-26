//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func assertAnalyticsExportOwnerRollback(t *testing.T, ctx context.Context,
	database *sql.DB, connection *sql.Conn,
) {
	t.Helper()
	transaction := analyticsExportTransaction(t, ctx, connection)
	_, err := transaction.ExecContext(ctx, `SELECT disposition
FROM analytics_dashboard_export_claim_v1('t_analytics_b',$1,$2,$3,$4)`,
		analyticsExportID, analyticsExportJob, analyticsExportLeaseA,
		analyticsExportAt.Add(time.Minute))
	if err == nil {
		_ = transaction.Rollback()
		t.Fatal("Analytics export claim accepted the wrong owner scope")
	}
	requireAnalyticsPGCode(t, err, "42501", "Analytics export owner failure")
	_ = transaction.Rollback()
	assertAnalyticsExportStateAndQuota(t, ctx, database, "queued", "none", 0, 0)
}

func claimAnalyticsExportConcurrently(t *testing.T, ctx context.Context,
	database *sql.DB, connection *sql.Conn,
) {
	t.Helper()
	first := analyticsExportTransaction(t, ctx, connection)
	var disposition, state string
	var generation int64
	err := first.QueryRowContext(ctx, `SELECT disposition,state,attempt_generation
FROM analytics_dashboard_export_claim_v1('t_analytics_a',$1,$2,$3,$4)`,
		analyticsExportID, analyticsExportJob, analyticsExportLeaseA,
		analyticsExportAt.Add(time.Minute)).Scan(&disposition, &state, &generation)
	if err != nil || disposition != "execute" || state != "running" || generation != 1 {
		_ = first.Rollback()
		t.Fatalf("first concurrent Analytics export claim = %s/%s/%d: %v",
			disposition, state, generation, err)
	}
	secondConnection := analyticsRuntimeConnection(t, ctx, database)
	second, err := secondConnection.BeginTx(ctx, nil)
	if err != nil {
		_ = first.Rollback()
		t.Fatal(err)
	}
	defer func() { _ = second.Rollback() }()
	if _, err = second.ExecContext(ctx, `SET LOCAL search_path=t_analytics_a,platform,pg_catalog;
SET LOCAL lock_timeout='500ms'`); err != nil {
		_ = first.Rollback()
		t.Fatal(err)
	}
	_, err = second.ExecContext(ctx, `SELECT disposition
FROM analytics_dashboard_export_claim_v1('t_analytics_a',$1,$2,$3,$4)`,
		analyticsExportID, analyticsExportJob, analyticsExportLeaseB,
		analyticsExportAt.Add(time.Minute))
	if err == nil {
		_ = first.Rollback()
		t.Fatal("concurrent Analytics export claim bypassed the owner row lock")
	}
	requireAnalyticsPGCode(t, err, "55P03", "Analytics export concurrent claim")
	_ = second.Rollback()
	if err = first.Commit(); err != nil {
		t.Fatal(err)
	}
	assertAnalyticsExportStateAndQuota(t, ctx, database,
		"running", "held", 0, 2147483648)
}

func assertAnalyticsExportStateAndQuota(t *testing.T, ctx context.Context,
	database *sql.DB, state, quotaState string, used, reserved int64,
) {
	t.Helper()
	var storedState, storedQuota string
	err := database.QueryRowContext(ctx, `SELECT state,quota_state
FROM t_analytics_a.analytics_dashboard_exports WHERE export_id=$1`,
		analyticsExportID).Scan(&storedState, &storedQuota)
	if err != nil || storedState != state || storedQuota != quotaState {
		t.Fatalf("Analytics export state/quota = %s/%s: %v", storedState, storedQuota, err)
	}
	assertAnalyticsExportQuota(t, ctx, database, used, reserved)
}

func assertAnalyticsExportQuota(t *testing.T, ctx context.Context,
	database *sql.DB, used, reserved int64,
) {
	t.Helper()
	var storedUsed, storedReserved int64
	err := database.QueryRowContext(ctx, `SELECT used_ciphertext_bytes,reserved_ciphertext_bytes
FROM platform.artifact_quota_counters WHERE tenant_id=$1`, analyticsTenantA).Scan(
		&storedUsed, &storedReserved)
	if err != nil || storedUsed != used || storedReserved != reserved {
		t.Fatalf("Analytics export tenant quota = %d/%d, want %d/%d: %v",
			storedUsed, storedReserved, used, reserved, err)
	}
}
