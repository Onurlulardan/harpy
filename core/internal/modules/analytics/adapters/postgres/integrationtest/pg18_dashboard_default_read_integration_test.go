//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

const analyticsDashboardDefaultOperation = "019b1000-0000-7000-8000-000000000059"

func assertAnalyticsDashboardDefaultRead(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	connection := analyticsRuntimeConnection(t, ctx, database)
	setAnalyticsDashboardDefault(t, ctx, connection)

	transaction, err := connection.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(ctx,
		`SET LOCAL search_path=t_analytics_a,platform,pg_catalog`); err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
	var dashboardID, name string
	err = transaction.QueryRowContext(ctx, `SELECT dashboard_id,name
FROM analytics_dashboard_get_default_v1('t_analytics_a',$1,'home')`,
		analyticsDashboardActor).Scan(&dashboardID, &name)
	if err != nil || dashboardID != analyticsDashboard || name == "" {
		_ = transaction.Rollback()
		t.Fatalf("read default dashboard in snapshot: id=%s name=%q err=%v",
			dashboardID, name, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}

	assertAnalyticsDefaultTableDenied(t, ctx, connection)
}

func setAnalyticsDashboardDefault(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) {
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
	var replayed bool
	err = transaction.QueryRowContext(ctx, `SELECT analytics_dashboard_set_default_v1(
't_analytics_a',$1,$2,'dashboard-default',decode(repeat('54',32),'hex'),
$2,$3,'home',clock_timestamp())`, analyticsDashboardDefaultOperation,
		analyticsDashboardActor, analyticsDashboard).Scan(&replayed)
	if err != nil || replayed {
		_ = transaction.Rollback()
		t.Fatalf("set default dashboard: replay=%v err=%v", replayed, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func assertAnalyticsDefaultTableDenied(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) {
	t.Helper()
	transaction, err := connection.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transaction.Rollback() }()
	if _, err = transaction.ExecContext(ctx,
		`SET LOCAL search_path=t_analytics_a,platform,pg_catalog`); err != nil {
		t.Fatal(err)
	}
	err = transaction.QueryRowContext(ctx,
		`SELECT count(*) FROM analytics_dashboard_preferences`).Scan(new(int))
	if err == nil {
		t.Fatal("runtime read Analytics dashboard preference table directly")
	}
	requireAnalyticsPGCode(t, err, "42501", "direct dashboard preference read")
}
