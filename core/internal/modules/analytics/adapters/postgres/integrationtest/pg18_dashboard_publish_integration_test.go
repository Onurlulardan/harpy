//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

const (
	analyticsDashboard       = "019b1000-0000-7000-8000-000000000051"
	analyticsDashboardDraft  = "019b1000-0000-7000-8000-000000000052"
	analyticsDashboardNext   = "019b1000-0000-7000-8000-000000000053"
	analyticsDashboardActor  = "019b1000-0000-7000-8000-000000000054"
	analyticsDashboardWidget = "019b1000-0000-7000-8000-000000000055"
)

func assertAnalyticsDashboardPublishing(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	connection := analyticsRuntimeConnection(t, ctx, database)
	replayed, err := createAnalyticsDashboard(ctx, connection)
	if err != nil || replayed {
		t.Fatalf("create Analytics dashboard: replay=%v err=%v", replayed, err)
	}
	replayed, err = publishAnalyticsDashboard(ctx, connection,
		"019b1000-0000-7000-8000-000000000057", "dashboard-publish", 1)
	if err != nil || replayed {
		t.Fatalf("publish Analytics dashboard: replay=%v err=%v", replayed, err)
	}
	replayed, err = publishAnalyticsDashboard(ctx, connection,
		"019b1000-0000-7000-8000-000000000057", "dashboard-publish", 1)
	if err != nil || !replayed {
		t.Fatalf("replay Analytics dashboard: replay=%v err=%v", replayed, err)
	}
	_, err = publishAnalyticsDashboard(ctx, connection,
		"019b1000-0000-7000-8000-000000000058", "dashboard-publish-stale", 1)
	if err == nil {
		t.Fatal("Analytics dashboard publish accepted stale root revision")
	}
	requireAnalyticsPGCode(t, err, "40001", "stale dashboard publish")
	assertPublishedDashboardImmutable(t, ctx, database)
}

func createAnalyticsDashboard(
	ctx context.Context, connection *sql.Conn,
) (bool, error) {
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = transaction.Rollback() }()
	if _, err = transaction.ExecContext(ctx,
		`SET LOCAL search_path=t_analytics_a,platform,pg_catalog`); err != nil {
		return false, err
	}
	var replayed bool
	err = transaction.QueryRowContext(ctx, `SELECT replayed FROM analytics_dashboard_create_v1(
't_analytics_a','019b1000-0000-7000-8000-000000000056',$1,'dashboard-create',
decode(repeat('51',32),'hex'),$2,$3,'Operations','tenant','{}'::bytea,
jsonb_build_array(jsonb_build_object('widget_id',$4::uuid,'position',0,
'query_version_id',$5::uuid,'parameter_document','7b7d','output_document','7b7d')),
clock_timestamp())`, analyticsDashboardActor, analyticsDashboard,
		analyticsDashboardDraft, analyticsDashboardWidget, analyticsPublishVersion).Scan(&replayed)
	if err != nil {
		return false, err
	}
	return replayed, transaction.Commit()
}

func publishAnalyticsDashboard(
	ctx context.Context, connection *sql.Conn, operationID, key string, rootRevision int,
) (bool, error) {
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = transaction.Rollback() }()
	if _, err = transaction.ExecContext(ctx,
		`SET LOCAL search_path=t_analytics_a,platform,pg_catalog`); err != nil {
		return false, err
	}
	var replayed bool
	err = transaction.QueryRowContext(ctx, `SELECT replayed FROM analytics_dashboard_publish_v1(
't_analytics_a',$1,$2,$3,decode(repeat('52',32),'hex'),$4,$5,$6,$7,1,
decode(repeat('53',32),'hex'),clock_timestamp())`, operationID, analyticsDashboardActor,
		key, analyticsDashboard, analyticsDashboardDraft, analyticsDashboardNext,
		rootRevision).Scan(&replayed)
	if err != nil {
		return false, err
	}
	return replayed, transaction.Commit()
}

func assertPublishedDashboardImmutable(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `UPDATE t_analytics_a.analytics_dashboard_versions
SET name='tampered',row_revision=row_revision+1,updated_at=clock_timestamp()
WHERE id=$1`, analyticsDashboardDraft)
	if err == nil {
		t.Fatal("published Analytics dashboard version was mutable")
	}
	requireAnalyticsPGCode(t, err, "P6102", "published dashboard mutation")
	_, err = database.ExecContext(ctx, `UPDATE t_analytics_a.analytics_dashboard_widgets
SET position=position WHERE dashboard_version_id=$1`, analyticsDashboardDraft)
	if err == nil {
		t.Fatal("published Analytics dashboard widget was mutable")
	}
	requireAnalyticsPGCode(t, err, "P6102", "published dashboard widget mutation")
}
