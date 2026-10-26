//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
)

func assertAnalyticsRetentionBoundaries(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	checks := []struct {
		table, column, id string
		want              int
	}{
		{"analytics_saved_query_versions", "id", retentionOldDraft, 0},
		{"analytics_saved_query_versions", "id", retentionRecentDraft, 1},
		{"analytics_saved_query_versions", "id", retentionHeldVersion, 1},
		{"analytics_execution_audits", "execution_id", "019b1000-0000-7000-8000-000000000121", 0},
		{"analytics_execution_audits", "execution_id", "019b1000-0000-7000-8000-000000000122", 1},
		{"analytics_publication_approvals", "release_id", "019b1000-0000-7000-8000-000000000123", 0},
		{"analytics_publication_approvals", "release_id", "019b1000-0000-7000-8000-000000000124", 1},
		{"analytics_operations", "operation_id", "019b1000-0000-7000-8000-000000000125", 0},
		{"analytics_operations", "operation_id", "019b1000-0000-7000-8000-000000000126", 1},
		{"analytics_capability_tokens", "release_id", "019b1000-0000-7000-8000-000000000128", 0},
		{"analytics_capability_tokens", "release_id", "019b1000-0000-7000-8000-000000000129", 1},
	}
	for _, check := range checks {
		assertAnalyticsRowCount(t, ctx, database, "t_analytics_a."+check.table,
			check.column, check.id, check.want)
	}
}

func assertAnalyticsRowCount(
	t *testing.T, ctx context.Context, database *sql.DB,
	table, column, id string, want int,
) {
	t.Helper()
	var count int
	query := fmt.Sprintf("SELECT count(*) FROM %s WHERE %s=$1", table, column)
	if err := database.QueryRowContext(ctx, query, id).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != want {
		t.Fatalf("Analytics retention %s.%s %s count=%d, want %d",
			table, column, id, count, want)
	}
}

func assertAnalyticsRetentionGuards(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	_, err := database.ExecContext(ctx,
		`DELETE FROM t_analytics_a.analytics_saved_query_versions WHERE id=$1`,
		analyticsPublishVersion)
	if err == nil {
		t.Fatal("published Analytics query was directly deleted")
	}
	requireAnalyticsPGCode(t, err, "P6102", "retention published mutation")
	_, err = database.ExecContext(ctx, `UPDATE t_analytics_a.analytics_retention_receipts
SET checkpoint_revision=checkpoint_revision+1 WHERE job_id=$1`,
		"019b1000-0000-7000-8000-000000000111")
	if err == nil {
		t.Fatal("Analytics retention receipt was mutable")
	}
	requireAnalyticsPGCode(t, err, "55000", "retention receipt mutation")
}

func assertAnalyticsRetentionBounded(
	t *testing.T, ctx context.Context, database *sql.DB, connection *sql.Conn,
) {
	t.Helper()
	_, err := database.ExecContext(ctx, `INSERT INTO t_analytics_a.analytics_operations
SELECT uuidv7(),$1,'query_delete',NULL,decode(repeat('91',32),'hex'),'query',$2,
1,decode(repeat('92',32),'hex'),clock_timestamp()-interval '8 years'
FROM generate_series(1,30)`, retentionActor,
		"019b1000-0000-7000-8000-000000000140")
	if err != nil {
		t.Fatalf("seed bounded Analytics retention: %v", err)
	}
	outcome := executeAnalyticsRetention(t, ctx, connection,
		"019b1000-0000-7000-8000-000000000141",
		"019b1000-0000-7000-8000-000000000142", 4)
	if outcome.totalPurged() > 4 || !outcome.hasRemaining {
		t.Fatalf("bounded Analytics retention = %+v", outcome)
	}
	var remaining int
	if err = database.QueryRowContext(ctx, `SELECT count(*) FROM t_analytics_a.analytics_operations
WHERE resource_id=$1`, "019b1000-0000-7000-8000-000000000140").Scan(&remaining); err != nil {
		t.Fatal(err)
	}
	if remaining != 29 {
		t.Fatalf("bounded Analytics retention remaining=%d, want 29", remaining)
	}
}
