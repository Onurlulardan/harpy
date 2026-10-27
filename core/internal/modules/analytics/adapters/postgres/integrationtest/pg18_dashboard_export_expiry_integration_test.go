//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"
)

const (
	analyticsExpiryExport      = "019b1000-0000-7000-8000-000000000081"
	analyticsExpiryRenderJob   = "019b1000-0000-7000-8000-000000000082"
	analyticsExpiryCleanupJob  = "019b1000-0000-7000-8000-000000000083"
	analyticsExpiryRenderLease = "019b1000-0000-7000-8000-000000000084"
	analyticsExpiryDeleteLease = "019b1000-0000-7000-8000-000000000085"
)

func assertAnalyticsExportAutomaticExpiry(t *testing.T, ctx context.Context,
	database *sql.DB, connection *sql.Conn,
) {
	t.Helper()
	expiresAt := analyticsExportAt.Add(24 * time.Hour)
	transaction := analyticsExportTransaction(t, ctx, connection)
	var exportID, state string
	err := transaction.QueryRowContext(ctx, `SELECT export_id,effective_state
FROM analytics_dashboard_export_submit_v2(
't_analytics_a',$1,$2,$3,$4,$5,$6,$7,$8,'csv','{}'::bytea,
1,1,1,100000,67108864,$9,$10)`,
		"019b1000-0000-7000-8000-000000000080", analyticsDashboardActor,
		"analytics-expiry-0001", bytes.Repeat([]byte{0xe1}, 32),
		analyticsExpiryExport, analyticsExpiryRenderJob, analyticsExpiryCleanupJob,
		analyticsDashboard, analyticsExportAt, expiresAt).Scan(&exportID, &state)
	if err != nil || exportID != analyticsExpiryExport || state != "queued" {
		_ = transaction.Rollback()
		t.Fatalf("submit expiring Analytics export = %s/%s: %v", exportID, state, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	var expiryJob string
	if err = database.QueryRowContext(ctx, `SELECT expiry_job_id
FROM t_analytics_a.analytics_dashboard_exports WHERE export_id=$1`,
		analyticsExpiryExport).Scan(&expiryJob); err != nil ||
		expiryJob != analyticsExpiryCleanupJob {
		t.Fatalf("pinned Analytics expiry job = %s: %v", expiryJob, err)
	}

	transaction = analyticsExportTransaction(t, ctx, connection)
	var disposition string
	var generation int64
	err = transaction.QueryRowContext(ctx, `SELECT disposition,attempt_generation
FROM analytics_dashboard_export_claim_v1('t_analytics_a',$1,$2,$3,$4)`,
		analyticsExpiryExport, analyticsExpiryRenderJob, analyticsExpiryRenderLease,
		analyticsExportAt.Add(time.Minute)).Scan(&disposition, &generation)
	if err != nil || disposition != "execute" || generation != 1 {
		_ = transaction.Rollback()
		t.Fatalf("claim expiring Analytics export = %s/%d: %v", disposition, generation, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}

	callAnalyticsExportBoolean(t, ctx, connection,
		`SELECT analytics_dashboard_export_effect_start_v1(
't_analytics_a',$1,$2,1,'analytics/expiry/object','analytics-kek-1',$3,$4,$5,$6,$7,
1024,128,1,'text/csv; charset=utf-8','expiry.csv',$8)`,
		analyticsExpiryExport, analyticsExpiryRenderLease,
		bytes.Repeat([]byte{0xe2}, 48), bytes.Repeat([]byte{0xe3}, 12),
		bytes.Repeat([]byte{0xe4}, 32), bytes.Repeat([]byte{0xe5}, 32),
		bytes.Repeat([]byte{0xe6}, 32), analyticsExportAt.Add(90*time.Second))
	transaction = analyticsExportTransaction(t, ctx, connection)
	err = transaction.QueryRowContext(ctx, `SELECT analytics_dashboard_export_complete_v1(
't_analytics_a',$1,$2,1,'version-expiry','provider-expiry',$3,$4)`,
		analyticsExpiryExport, analyticsExpiryRenderLease,
		analyticsExportAt.Add(100*time.Second),
		analyticsExportAt.Add(110*time.Second)).Scan(&state)
	if err != nil || state != "ready" {
		_ = transaction.Rollback()
		t.Fatalf("complete expiring Analytics export = %s: %v", state, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	assertAnalyticsExportQuota(t, ctx, database, 1024, 0)

	transaction = analyticsExportTransaction(t, ctx, connection)
	_, err = transaction.ExecContext(ctx, `SELECT disposition
FROM analytics_dashboard_export_delete_claim_v1('t_analytics_a',$1,$2,$3,$4)`,
		analyticsExpiryExport, analyticsExpiryCleanupJob, analyticsExpiryDeleteLease,
		expiresAt.Add(-time.Second))
	if err == nil {
		_ = transaction.Rollback()
		t.Fatal("Analytics expiry cleanup claimed the artifact before expires_at")
	}
	requireAnalyticsPGCode(t, err, "55P03", "Analytics export early expiry")
	_ = transaction.Rollback()
	assertAnalyticsExpiryState(t, ctx, database, "ready", "used")

	transaction = analyticsExportTransaction(t, ctx, connection)
	var objectVersion string
	err = transaction.QueryRowContext(ctx, `SELECT disposition,state,object_version,
delete_attempt_generation FROM analytics_dashboard_export_delete_claim_v1(
't_analytics_a',$1,$2,$3,$4)`, analyticsExpiryExport,
		analyticsExpiryCleanupJob, analyticsExpiryDeleteLease, expiresAt).Scan(
		&disposition, &state, &objectVersion, &generation)
	if err != nil || disposition != "execute" || state != "deleting" ||
		objectVersion != "version-expiry" || generation != 1 {
		_ = transaction.Rollback()
		t.Fatalf("claim Analytics expiry cleanup = %s/%s/%s/%d: %v",
			disposition, state, objectVersion, generation, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	callAnalyticsExportBoolean(t, ctx, connection,
		`SELECT analytics_dashboard_export_delete_complete_v1(
't_analytics_a',$1,$2,$3,1,'version-expiry','provider-expiry-delete',$4,$5,$6,$7)`,
		analyticsExpiryExport, analyticsExpiryCleanupJob, analyticsExpiryDeleteLease,
		bytes.Repeat([]byte{0xe7}, 32), bytes.Repeat([]byte{0xe8}, 32),
		expiresAt.Add(time.Minute), expiresAt.Add(2*time.Minute))
	assertAnalyticsExpiryState(t, ctx, database, "deleted", "released")
	assertAnalyticsExportQuota(t, ctx, database, 0, 0)

	transaction = analyticsExportTransaction(t, ctx, connection)
	var schedule, replayed bool
	err = transaction.QueryRowContext(ctx, `SELECT replayed,schedule_delete
FROM analytics_dashboard_export_delete_v1('t_analytics_a',$1,$2,$3,$4,$5,$6,$7,$8)`,
		"019b1000-0000-7000-8000-000000000086",
		"019b1000-0000-7000-8000-000000000087", analyticsDashboardActor,
		"analytics-delete-after-expiry", bytes.Repeat([]byte{0xe9}, 32),
		analyticsDashboard, analyticsExpiryExport, expiresAt.Add(3*time.Minute)).Scan(
		&replayed, &schedule)
	if err != nil || replayed || schedule {
		_ = transaction.Rollback()
		t.Fatalf("delete after automatic expiry = %v/%v: %v", replayed, schedule, err)
	}
	_ = transaction.Rollback()
}

func assertAnalyticsExpiryState(t *testing.T, ctx context.Context,
	database *sql.DB, state, quota string,
) {
	t.Helper()
	var storedState, storedQuota string
	err := database.QueryRowContext(ctx, `SELECT state,quota_state
FROM t_analytics_a.analytics_dashboard_exports WHERE export_id=$1`,
		analyticsExpiryExport).Scan(&storedState, &storedQuota)
	if err != nil || storedState != state || storedQuota != quota {
		t.Fatalf("expiring Analytics export state = %s/%s: %v",
			storedState, storedQuota, err)
	}
}
