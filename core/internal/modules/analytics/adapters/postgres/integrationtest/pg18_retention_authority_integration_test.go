//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func registerAnalyticsRetentionReference(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) {
	t.Helper()
	analyticsRetentionRuntimeExec(t, ctx, connection, `SELECT * FROM analytics_reference_register_v1(
't_analytics_a'::name,$1::uuid,$2::uuid,'query_version',$3::uuid,
'workflow','definition',$4::uuid,$5::uuid,'read',clock_timestamp()-interval '450 days')`,
		retentionReference, retentionActor, retentionHeldVersion,
		"019b1000-0000-7000-8000-000000000108",
		"019b1000-0000-7000-8000-000000000109")
}

func releaseAnalyticsRetentionReference(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) {
	t.Helper()
	analyticsRetentionRuntimeExec(t, ctx, connection,
		`SELECT * FROM analytics_reference_release_v1(
't_analytics_a'::name,$1::uuid,$2::uuid,clock_timestamp())`,
		retentionReference, retentionActor)
}

func placeAnalyticsRetentionHold(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) {
	t.Helper()
	analyticsRetentionRuntimeExec(t, ctx, connection,
		`SELECT * FROM analytics_legal_hold_place_v1(
't_analytics_a'::name,$1::uuid,$2::uuid,decode(repeat('aa',32),'hex'),clock_timestamp())`,
		retentionHold, retentionActor)
}

func releaseAnalyticsRetentionHold(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) {
	t.Helper()
	analyticsRetentionRuntimeExec(t, ctx, connection,
		`SELECT * FROM analytics_legal_hold_release_v1(
't_analytics_a'::name,$1::uuid,$2::uuid,1,clock_timestamp())`,
		retentionHold, retentionActor)
}

func retireAnalyticsReferencedVersion(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin referenced Analytics retirement: %v", err)
	}
	defer func() { _ = transaction.Rollback() }()
	_, err = transaction.ExecContext(ctx, `UPDATE t_analytics_a.analytics_saved_queries
SET lifecycle_state='retired',row_revision=row_revision+1,
updated_at=clock_timestamp()-interval '401 days'
WHERE id=(SELECT query_id FROM t_analytics_a.analytics_saved_query_versions WHERE id=$1)`,
		retentionHeldVersion)
	if err == nil {
		_, err = transaction.ExecContext(ctx, `UPDATE t_analytics_a.analytics_saved_query_versions
SET version_state='retired',row_revision=row_revision+1,
updated_at=clock_timestamp()-interval '401 days'
WHERE id=$1`, retentionHeldVersion)
	}
	if err != nil {
		t.Fatalf("retire referenced Analytics version: %v", err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatalf("commit referenced Analytics retirement: %v", err)
	}
}
