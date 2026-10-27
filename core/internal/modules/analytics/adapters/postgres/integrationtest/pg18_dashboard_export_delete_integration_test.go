//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"
)

func assertAnalyticsExportContent(t *testing.T, ctx context.Context, connection *sql.Conn) {
	t.Helper()
	transaction := analyticsExportTransaction(t, ctx, connection)
	defer func() { _ = transaction.Rollback() }()
	var objectVersion, contentType string
	err := transaction.QueryRowContext(ctx, `SELECT object_version,content_type
FROM analytics_dashboard_export_content_v1('t_analytics_a',$1,$2,$3,$4)`,
		analyticsDashboardActor, analyticsDashboard, analyticsExportID,
		analyticsExportAt.Add(10*time.Minute)).Scan(&objectVersion, &contentType)
	if err != nil || objectVersion != "version-2" || contentType != "text/csv; charset=utf-8" {
		t.Fatalf("Analytics export content = %s/%s: %v", objectVersion, contentType, err)
	}
}

func assertAnalyticsExportTenantIsolation(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) {
	t.Helper()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transaction.Rollback() }()
	if _, err = transaction.ExecContext(ctx,
		`SET LOCAL search_path=t_analytics_b,platform,pg_catalog`); err != nil {
		t.Fatal(err)
	}
	_, err = transaction.ExecContext(ctx, `SELECT export_id
FROM analytics_dashboard_export_snapshot_v1('t_analytics_b',$1,$2,$3,$4)`,
		analyticsDashboardActor, analyticsDashboard, analyticsExportID,
		analyticsExportAt.Add(10*time.Minute))
	if err == nil {
		t.Fatal("tenant B read tenant A Analytics export")
	}
	requireAnalyticsPGCode(t, err, "P6102", "Analytics export tenant isolation")
}

func deleteAnalyticsExport(t *testing.T, ctx context.Context, connection *sql.Conn) {
	t.Helper()
	transaction := analyticsExportTransaction(t, ctx, connection)
	defer func() { _ = transaction.Rollback() }()
	var state string
	var schedule bool
	err := transaction.QueryRowContext(ctx, `SELECT effective_state,schedule_delete
FROM analytics_dashboard_export_delete_v1('t_analytics_a',$1,$2,$3,$4,$5,$6,$7,$8)`,
		"019b1000-0000-7000-8000-00000000006c", analyticsExportDeleteJob,
		analyticsDashboardActor, "analytics-delete-0001", bytes.Repeat([]byte{0xc1}, 32),
		analyticsDashboard, analyticsExportID, analyticsExportAt.Add(11*time.Minute)).Scan(
		&state, &schedule)
	if err != nil || state != "deleting" || !schedule {
		t.Fatalf("delete Analytics export = %s/%v: %v", state, schedule, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func completeAnalyticsExportDelete(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) {
	t.Helper()
	transaction := analyticsExportTransaction(t, ctx, connection)
	defer func() { _ = transaction.Rollback() }()
	var disposition, objectVersion string
	var deleteGeneration int64
	err := transaction.QueryRowContext(ctx, `SELECT disposition,object_version,
delete_attempt_generation FROM analytics_dashboard_export_delete_claim_v1(
't_analytics_a',$1,$2,$3,$4)`, analyticsExportID, analyticsExportDeleteJob,
		analyticsExportLeaseA, analyticsExportAt.Add(12*time.Minute)).Scan(
		&disposition, &objectVersion, &deleteGeneration)
	if err != nil || disposition != "execute" || objectVersion != "version-2" ||
		deleteGeneration != 1 {
		t.Fatalf("claim Analytics export delete = %s/%s/%d: %v",
			disposition, objectVersion, deleteGeneration, err)
	}
	var persisted bool
	err = transaction.QueryRowContext(ctx, `SELECT analytics_dashboard_export_delete_complete_v1(
't_analytics_a',$1,$2,$3,1,'version-2','provider-delete-1',$4,$5,$6,$7)`,
		analyticsExportID, analyticsExportDeleteJob, analyticsExportLeaseA,
		bytes.Repeat([]byte{0xd1}, 32), bytes.Repeat([]byte{0xd2}, 32),
		analyticsExportAt.Add(13*time.Minute), analyticsExportAt.Add(14*time.Minute)).Scan(&persisted)
	if err != nil || !persisted {
		t.Fatalf("complete Analytics export delete = %v: %v", persisted, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func assertAnalyticsExportTableDenied(t *testing.T, ctx context.Context,
	connection *sql.Conn,
) {
	t.Helper()
	if _, err := connection.ExecContext(ctx,
		`SELECT count(*) FROM t_analytics_a.analytics_dashboard_exports`); err == nil {
		t.Fatal("harpy_runtime read Analytics export storage directly")
	} else {
		requireAnalyticsPGCode(t, err, "42501", "Analytics export storage ACL")
	}
}
