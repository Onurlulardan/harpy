//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"
)

func assertAnalyticsDashboardExportExpiryLifecycle(t *testing.T,
	ctx context.Context, database *sql.DB, connection *sql.Conn, expiresAt time.Time,
) {
	t.Helper()
	claim := claimAnalyticsDashboardExportPlatform(t, ctx, connection,
		analyticsExpiryExportV2, analyticsExpiryRenderJobV2,
		analyticsExpiryRenderLeaseV2, 1)
	checkpointAt := checkpointAnalyticsDashboardExportPlatform(t, ctx, connection,
		analyticsExpiryExportV2, analyticsExpiryRenderJobV2, claim, 1,
		analyticsExpiryRenderLeaseV2)
	effect := startAnalyticsDashboardExportPlatformEffect(t, ctx, connection,
		analyticsExpiryExportV2, analyticsExpiryRenderJobV2,
		analyticsExpiryRenderLeaseV2, analyticsExpiryRenderEffect, claim, 1, 0xe1,
		checkpointAt.Add(time.Millisecond))
	succeedAnalyticsDashboardExportPlatform(t, ctx, connection,
		analyticsExpiryExportV2, analyticsExpiryRenderJobV2,
		analyticsExpiryRenderLeaseV2, analyticsExpiryReceipt, claim, effect, 1)
	assertAnalyticsExportQuota(t, ctx, database, 2048, 0)
	renderRoot := analyticsDashboardExportPlatformRootSnapshot(t, ctx, database,
		analyticsExpiryRenderJobV2)
	waitUntilAnalyticsDashboardExportExpiry(t, ctx, expiresAt)
	observedAt := time.Now().UTC().Truncate(time.Microsecond)
	before := analyticsDashboardExportLifecycleSnapshot(t, ctx, database,
		analyticsExpiryExportV2, analyticsExpiryRenderJobV2)
	state, cleanup := controlAnalyticsDashboardExportExpiry(t, ctx, connection, observedAt)
	if state != "ready" || !cleanup {
		t.Fatalf("Platform-first expiry control = %s/%v", state, cleanup)
	}
	after := analyticsDashboardExportLifecycleSnapshot(t, ctx, database,
		analyticsExpiryExportV2, analyticsExpiryRenderJobV2)
	state, cleanup = controlAnalyticsDashboardExportExpiry(t, ctx, connection, observedAt)
	if state != "ready" || !cleanup || before == after || after !=
		analyticsDashboardExportLifecycleSnapshot(t, ctx, database,
			analyticsExpiryExportV2, analyticsExpiryRenderJobV2) {
		t.Fatal("Platform-first expiry replay changed exact evidence")
	}
	completeAnalyticsDashboardExportMaintenanceDelete(t, ctx, connection,
		database, analyticsExpiryExportV2, analyticsExpiryCleanupJobV2,
		analyticsExpiryAttemptV2, analyticsExpiryLeaseV2,
		"analytics.dashboard.export.expire",
		observedAt.Add(time.Millisecond))
	if renderRoot != analyticsDashboardExportPlatformRootSnapshot(t, ctx, database,
		analyticsExpiryRenderJobV2) {
		t.Fatal("expiry cleanup mutated the terminal render Platform root")
	}
	assertAnalyticsDashboardExportStoredState(t, ctx, database,
		analyticsExpiryExportV2, "deleted", "released")
	assertAnalyticsExportQuota(t, ctx, database, 1024, 0)
}

func controlAnalyticsDashboardExportExpiry(t *testing.T, ctx context.Context,
	connection *sql.Conn, observedAt time.Time,
) (string, bool) {
	t.Helper()
	transaction := analyticsExportTransaction(t, ctx, connection)
	defer func() { _ = transaction.Rollback() }()
	var state string
	var cleanup bool
	err := transaction.QueryRowContext(ctx, `SELECT state,cleanup_required FROM
analytics_dashboard_export_expiry_control_v1(
't_analytics_a',$1,$2,$3,$4)`, analyticsTenantA, analyticsExpiryExportV2,
		analyticsExpiryCleanupJobV2, observedAt).Scan(&state, &cleanup)
	if err != nil {
		t.Fatalf("Platform-first expiry control: %v", err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return state, cleanup
}

func waitUntilAnalyticsDashboardExportExpiry(t *testing.T, ctx context.Context,
	expiresAt time.Time,
) {
	t.Helper()
	delay := time.Until(expiresAt.Add(10 * time.Millisecond))
	if delay <= 0 {
		return
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
}

func deleteAnalyticsDashboardExportPlatformFirst(t *testing.T, ctx context.Context,
	database *sql.DB, connection *sql.Conn,
) {
	t.Helper()
	renderRoot := analyticsDashboardExportPlatformRootSnapshot(t, ctx, database,
		analyticsExportJob)
	transaction := analyticsExportTransaction(t, ctx, connection)
	defer func() { _ = transaction.Rollback() }()
	var state string
	var schedule bool
	err := transaction.QueryRowContext(ctx, `SELECT effective_state,schedule_delete
FROM analytics_dashboard_export_delete_v1('t_analytics_a',$1,$2,$3,$4,$5,$6,$7,$8)`,
		"019b1000-0000-7000-8000-00000000006c", analyticsExportDeleteJob,
		analyticsDashboardActor, "analytics-delete-platform-v1",
		bytes.Repeat([]byte{0xc1}, 32), analyticsDashboard, analyticsExportID,
		time.Now().UTC().Truncate(time.Microsecond)).Scan(&state, &schedule)
	if err != nil || state != "deleting" || !schedule {
		t.Fatalf("Platform-first delete request = %s/%v: %v", state, schedule, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	completeAnalyticsDashboardExportMaintenanceDelete(t, ctx, connection,
		database, analyticsExportID, analyticsExportDeleteJob,
		analyticsDeleteAttemptV2, analyticsDeleteLeaseV2,
		"analytics.dashboard.export.delete",
		time.Now().UTC().Truncate(time.Microsecond))
	if renderRoot != analyticsDashboardExportPlatformRootSnapshot(t, ctx, database,
		analyticsExportJob) {
		t.Fatal("manual cleanup mutated the terminal render Platform root")
	}
	assertAnalyticsDashboardExportStoredState(t, ctx, database, analyticsExportID,
		"deleted", "released")
	assertAnalyticsExportQuota(t, ctx, database, 0, 0)
}

func completeAnalyticsDashboardExportMaintenanceDelete(t *testing.T,
	ctx context.Context, connection *sql.Conn, database *sql.DB, exportID, jobID,
	attemptID, leaseOwner, kind string, observedAt time.Time,
) {
	t.Helper()
	effectID := analyticsDashboardExportMaintenanceEffectID(t, kind, attemptID)
	transaction := analyticsExportTransaction(t, ctx, connection)
	defer func() { _ = transaction.Rollback() }()
	var disposition, state, objectVersion string
	var generation int64
	err := transaction.QueryRowContext(ctx, `SELECT disposition,state,object_version,
delete_attempt_generation FROM analytics_dashboard_export_delete_claim_v1(
't_analytics_a',$1,$2,$3,$4)`, exportID, jobID, effectID, observedAt).Scan(
		&disposition, &state, &objectVersion, &generation)
	if err != nil || disposition != "execute" || state != "deleting" ||
		objectVersion != "version-2" || generation != 1 {
		t.Fatalf("fenced delete claim %s = %s/%s/%s/%d: %v", exportID,
			disposition, state, objectVersion, generation, err)
	}
	var persisted bool
	err = transaction.QueryRowContext(ctx, `SELECT
analytics_dashboard_export_delete_complete_v1(
't_analytics_a',$1,$2,$3,1,'version-2','provider-delete-v2',$4,$5,$6,$7)`,
		exportID, jobID, effectID, bytes.Repeat([]byte{0xf2}, 32),
		bytes.Repeat([]byte{0xf3}, 32), observedAt.Add(time.Millisecond),
		observedAt.Add(2*time.Millisecond)).Scan(&persisted)
	if err != nil || !persisted {
		t.Fatalf("fenced delete complete %s = %v: %v", exportID, persisted, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	beforeReplay := analyticsDashboardExportLifecycleSnapshot(t, ctx, database,
		exportID, jobID)
	replay := analyticsExportTransaction(t, ctx, connection)
	err = replay.QueryRowContext(ctx, `SELECT disposition,state FROM
analytics_dashboard_export_delete_claim_v1('t_analytics_a',$1,$2,$3,$4)`,
		exportID, jobID, effectID, observedAt.Add(3*time.Millisecond)).Scan(
		&disposition, &state)
	if err != nil || disposition != "terminal" || state != "deleted" {
		t.Fatalf("fenced delete production replay %s = %s/%s: %v",
			exportID, disposition, state, err)
	}
	if err = replay.Commit(); err != nil {
		t.Fatal(err)
	}
	if beforeReplay != analyticsDashboardExportLifecycleSnapshot(t, ctx, database,
		exportID, jobID) {
		t.Fatalf("fenced delete production replay %s mutated tenant, Platform, quota, or capacity", exportID)
	}
	completeAnalyticsDashboardExportPlatformMaintenance(t, ctx, database,
		connection, exportID, jobID, attemptID, leaseOwner, kind, "deleted")
}
