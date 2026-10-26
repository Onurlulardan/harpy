//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func analyticsExportTransaction(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) *sql.Tx {
	t.Helper()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(ctx,
		`SET LOCAL search_path=t_analytics_a,platform,pg_catalog`); err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
	return transaction
}

func claimAnalyticsExport(t *testing.T, ctx context.Context, connection *sql.Conn,
	leaseOwner, wantDisposition string, wantGeneration int64, observedAt time.Time,
) {
	t.Helper()
	transaction := analyticsExportTransaction(t, ctx, connection)
	defer func() { _ = transaction.Rollback() }()
	var disposition, state string
	var generation int64
	err := transaction.QueryRowContext(ctx, `SELECT disposition,state,attempt_generation
FROM analytics_dashboard_export_claim_v1('t_analytics_a',$1,$2,$3,$4)`,
		analyticsExportID, analyticsExportJob, leaseOwner, observedAt).Scan(
		&disposition, &state, &generation)
	if err != nil || disposition != wantDisposition || state != "running" ||
		generation != wantGeneration {
		t.Fatalf("claim Analytics export = %s/%s/%d: %v",
			disposition, state, generation, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func startAnalyticsExportEffect(t *testing.T, ctx context.Context, connection *sql.Conn,
	leaseOwner string, generation int64, startedAt time.Time,
) {
	t.Helper()
	callAnalyticsExportBoolean(t, ctx, connection,
		`SELECT analytics_dashboard_export_effect_start_v1(
't_analytics_a',$1,$2,$3,$4,'analytics-kek-1',$5,$6,$7,$8,$9,
1024,128,1,'text/csv; charset=utf-8','dashboard.csv',$10)`,
		analyticsExportID, leaseOwner, generation,
		fmt.Sprintf("analytics/dashboard/export/generation-%d", generation),
		bytes.Repeat([]byte{0xb1}, 48), bytes.Repeat([]byte{0xb2}, 12),
		bytes.Repeat([]byte{0xb3}, 32), bytes.Repeat([]byte{0xb4}, 32),
		bytes.Repeat([]byte{0xb5}, 32), startedAt)
}

func completeAnalyticsExport(t *testing.T, ctx context.Context, connection *sql.Conn) {
	t.Helper()
	transaction := analyticsExportTransaction(t, ctx, connection)
	defer func() { _ = transaction.Rollback() }()
	var state string
	err := transaction.QueryRowContext(ctx, `SELECT analytics_dashboard_export_complete_v1(
't_analytics_a',$1,$2,2,'version-2','provider-upload-2',$3,$4)`,
		analyticsExportID, analyticsExportLeaseB, analyticsExportAt.Add(8*time.Minute),
		analyticsExportAt.Add(9*time.Minute)).Scan(&state)
	if err != nil || state != "ready" {
		t.Fatalf("complete Analytics export = %s: %v", state, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func callAnalyticsExportBoolean(t *testing.T, ctx context.Context,
	connection *sql.Conn, statement string, arguments ...any,
) {
	t.Helper()
	transaction := analyticsExportTransaction(t, ctx, connection)
	defer func() { _ = transaction.Rollback() }()
	var persisted bool
	if err := transaction.QueryRowContext(ctx, statement, arguments...).Scan(&persisted); err != nil ||
		!persisted {
		t.Fatalf("Analytics export mutation = %v: %v", persisted, err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}
