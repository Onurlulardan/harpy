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
	analyticsExportID        = "019b1000-0000-7000-8000-000000000061"
	analyticsExportJob       = "019b1000-0000-7000-8000-000000000062"
	analyticsExportExpiryJob = "019b1000-0000-7000-8000-00000000006e"
	analyticsExportLeaseA    = "019b1000-0000-7000-8000-000000000063"
	analyticsExportLeaseB    = "019b1000-0000-7000-8000-000000000064"
	analyticsExportDeleteJob = "019b1000-0000-7000-8000-000000000065"
	analyticsExportKey       = "analytics-export-0001"
)

var analyticsExportAt = time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)

func assertAnalyticsDashboardExport(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	prepareAnalyticsDashboardExportPlatformLifecycle(t, ctx, database,
		analyticsDashboardExportOrigin(t))
	connection := analyticsRuntimeConnection(t, ctx, database)
	analyticsExportAt = time.Now().UTC().Add(-2 * time.Second).Truncate(time.Microsecond)
	submitAnalyticsExport(t, ctx, connection)
	submitAnalyticsDashboardExportFixture(t, ctx, connection,
		analyticsFailureOperation, analyticsFailureExport, analyticsFailureJob,
		analyticsFailureExpiryJob, "analytics-failure-platform-v1", 0xa3,
		analyticsExportAt, analyticsExportAt.Add(24*time.Hour))
	expiresAt := time.Now().UTC().Add(15 * time.Second).Truncate(time.Microsecond)
	submitAnalyticsDashboardExportFixture(t, ctx, connection,
		analyticsExpiryOperationV2, analyticsExpiryExportV2,
		analyticsExpiryRenderJobV2, analyticsExpiryCleanupJobV2,
		"analytics-expiry-platform-v1", 0xa4, analyticsExportAt, expiresAt)
	assertAnalyticsExportReplay(t, ctx, connection)
	assertAnalyticsDashboardExportLegacyRenderDenied(t, ctx, database, connection)
	claim := claimAnalyticsDashboardExportPlatform(t, ctx, connection,
		analyticsExportID, analyticsExportJob, analyticsExportLeaseA, 1)
	assertAnalyticsDashboardExportPlatformClaimReplay(t, ctx, database, connection, claim)
	assertAnalyticsExportQuota(t, ctx, database, 0, 2147483648)
	checkpointAt := checkpointAnalyticsDashboardExportPlatform(t, ctx, connection,
		analyticsExportID, analyticsExportJob, claim, 1, analyticsExportLeaseA)
	effect := startAnalyticsDashboardExportPlatformEffect(t, ctx, connection,
		analyticsExportID, analyticsExportJob, analyticsExportLeaseA,
		analyticsMainEffect, claim, 1, 0xb1, checkpointAt.Add(time.Millisecond))
	assertAnalyticsDashboardExportEffectReplay(t, ctx, database, connection, claim, effect)
	succeedAnalyticsDashboardExportPlatform(t, ctx, connection, analyticsExportID,
		analyticsExportJob, analyticsExportLeaseA, analyticsMainReceipt, claim, effect, 1)
	assertAnalyticsDashboardExportSuccessReplay(t, ctx, database, connection, claim, effect)
	assertAnalyticsExportQuota(t, ctx, database, 1024, 0)
	assertAnalyticsDashboardExportFailureLifecycle(t, ctx, database, connection)
	assertAnalyticsDashboardExportExpiryLifecycle(t, ctx, database, connection, expiresAt)
	assertAnalyticsExportContent(t, ctx, connection)
	assertAnalyticsExportTenantIsolation(t, ctx, connection)
	deleteAnalyticsDashboardExportPlatformFirst(t, ctx, database, connection)
	assertAnalyticsExportTableDenied(t, ctx, connection)
}

func submitAnalyticsExport(t *testing.T, ctx context.Context, connection *sql.Conn) {
	t.Helper()
	transaction := analyticsExportTransaction(t, ctx, connection)
	defer func() { _ = transaction.Rollback() }()
	var exportID, state string
	var replayed bool
	err := transaction.QueryRowContext(ctx, `SELECT export_id,effective_state,replayed
FROM analytics_dashboard_export_submit_v2('t_analytics_a',$1,$2,$3,$4,$5,$6,$7,$8,
'csv','{}'::bytea,1,1,1,100000,67108864,$9,$10)`,
		"019b1000-0000-7000-8000-000000000060", analyticsDashboardActor,
		analyticsExportKey, bytes.Repeat([]byte{0xa1}, 32), analyticsExportID,
		analyticsExportJob, analyticsExportExpiryJob, analyticsDashboard, analyticsExportAt,
		analyticsExportAt.Add(24*time.Hour)).Scan(&exportID, &state, &replayed)
	if err != nil || exportID != analyticsExportID || state != "queued" || replayed {
		t.Fatalf("submit Analytics export = %s/%s/%v: %v", exportID, state, replayed, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func assertAnalyticsExportReplay(t *testing.T, ctx context.Context, connection *sql.Conn) {
	t.Helper()
	transaction := analyticsExportTransaction(t, ctx, connection)
	defer func() { _ = transaction.Rollback() }()
	var replayed bool
	err := transaction.QueryRowContext(ctx, `SELECT replayed
FROM analytics_dashboard_export_submit_v2('t_analytics_a',$1,$2,$3,$4,$5,$6,$7,$8,
'csv','{}'::bytea,1,1,1,100000,67108864,$9,$10)`,
		"019b1000-0000-7000-8000-000000000066", analyticsDashboardActor,
		analyticsExportKey, bytes.Repeat([]byte{0xa1}, 32),
		"019b1000-0000-7000-8000-000000000067",
		"019b1000-0000-7000-8000-000000000068",
		"019b1000-0000-7000-8000-00000000006f", analyticsDashboard,
		analyticsExportAt.Add(time.Minute), analyticsExportAt.Add(24*time.Hour)).Scan(&replayed)
	if err != nil || !replayed {
		t.Fatalf("replay Analytics export = %v: %v", replayed, err)
	}
	_, err = transaction.ExecContext(ctx, `SELECT export_id
FROM analytics_dashboard_export_submit_v2('t_analytics_a',$1,$2,$3,$4,$5,$6,$7,$8,
'csv','{}'::bytea,1,1,1,100000,67108864,$9,$10)`,
		"019b1000-0000-7000-8000-000000000069", analyticsDashboardActor,
		analyticsExportKey, bytes.Repeat([]byte{0xa2}, 32),
		"019b1000-0000-7000-8000-00000000006a",
		"019b1000-0000-7000-8000-00000000006b",
		"019b1000-0000-7000-8000-000000000070", analyticsDashboard,
		analyticsExportAt.Add(time.Minute), analyticsExportAt.Add(24*time.Hour))
	if err == nil {
		t.Fatal("Analytics export accepted reused key with a different digest")
	}
	requireAnalyticsPGCode(t, err, "P6103", "Analytics export idempotency mismatch")
}
