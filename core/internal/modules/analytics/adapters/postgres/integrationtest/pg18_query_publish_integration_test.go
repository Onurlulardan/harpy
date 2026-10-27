//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

const (
	analyticsPublishQuery     = "019b1000-0000-7000-8000-000000000023"
	analyticsPublishVersion   = "019b1000-0000-7000-8000-000000000024"
	analyticsPublishNextDraft = "019b1000-0000-7000-8000-000000000026"
	analyticsPublishActor     = "019b1000-0000-7000-8000-000000000022"
)

func assertAnalyticsQueryPublishing(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	connection := analyticsRuntimeConnection(t, ctx, database)
	replayed, err := publishAnalyticsQuery(ctx, connection,
		"019b1000-0000-7000-8000-000000000025", "query-publish", 1)
	if err != nil || replayed {
		t.Fatalf("publish Analytics query: replay=%v err=%v", replayed, err)
	}
	assertLatestPublishedQuery(t, ctx, connection)
	replayed, err = publishAnalyticsQuery(ctx, connection,
		"019b1000-0000-7000-8000-000000000025", "query-publish", 1)
	if err != nil || !replayed {
		t.Fatalf("replay Analytics query publish: replay=%v err=%v", replayed, err)
	}
	_, err = publishAnalyticsQuery(ctx, connection,
		"019b1000-0000-7000-8000-000000000027", "query-publish-stale", 1)
	if err == nil {
		t.Fatal("Analytics query publish accepted stale root revision")
	}
	requireAnalyticsPGCode(t, err, "40001", "stale query publish")
	assertPublishedQueryImmutable(t, ctx, database)
}

func assertLatestPublishedQuery(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) {
	t.Helper()
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transaction.Rollback() }()
	if _, err = transaction.ExecContext(ctx,
		`SET LOCAL search_path=t_analytics_a,platform,pg_catalog`); err != nil {
		t.Fatal(err)
	}
	var queryID, versionID, versionState string
	if err = transaction.QueryRowContext(ctx, `
SELECT query_id,version_id,version_state
FROM analytics_query_latest_published_v1('t_analytics_a',$1)`,
		analyticsPublishQuery).Scan(&queryID, &versionID, &versionState); err != nil {
		t.Fatal(err)
	}
	if queryID != analyticsPublishQuery || versionID != analyticsPublishVersion ||
		versionState != "published" {
		t.Fatalf("latest published query = %s/%s/%s", queryID, versionID, versionState)
	}
}

func publishAnalyticsQuery(
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
	err = transaction.QueryRowContext(ctx, `SELECT replayed FROM analytics_query_publish_v1(
't_analytics_a',$1,$2,$3,decode(repeat('32',32),'hex'),$4,$5,$6,$7,1,
'{}'::bytea,'{}'::bytea,'{}'::bytea,
jsonb_build_array(jsonb_build_object(
 'owner_domain','forms','resource_kind','form','resource_id',$8::uuid,
 'resource_version_id',$9::uuid,'owner_generation',1,'catalog_version',1,
 'query_contract_version',1,'policy_version',1,
 'data_revision_capability','record_revision','dependency_sha256',repeat('44',32))),
decode(repeat('45',32),'hex'),'revision_bound',clock_timestamp())`,
		operationID, analyticsPublishActor, key, analyticsPublishQuery,
		analyticsPublishVersion, analyticsPublishNextDraft, rootRevision,
		"019b1000-0000-7000-8000-000000000041",
		"019b1000-0000-7000-8000-000000000042").Scan(&replayed)
	if err != nil {
		return false, err
	}
	return replayed, transaction.Commit()
}

func assertPublishedQueryImmutable(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `UPDATE t_analytics_a.analytics_saved_query_versions
SET name='tampered',row_revision=row_revision+1,updated_at=clock_timestamp()
WHERE id=$1`, analyticsPublishVersion)
	if err == nil {
		t.Fatal("published Analytics query version was mutable")
	}
	requireAnalyticsPGCode(t, err, "P6102", "published query mutation")
	_, err = database.ExecContext(ctx, `UPDATE t_analytics_a.analytics_query_dependencies
SET owner_generation=2 WHERE query_version_id=$1`, analyticsPublishVersion)
	if err == nil {
		t.Fatal("published Analytics query dependency was mutable")
	}
	requireAnalyticsPGCode(t, err, "55000", "published dependency mutation")
}
