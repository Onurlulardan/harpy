//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

const (
	retentionActor       = "019b1000-0000-7000-8000-000000000100"
	retentionOldDraft    = "019b1000-0000-7000-8000-000000000102"
	retentionRecentDraft = "019b1000-0000-7000-8000-000000000104"
	retentionHeldVersion = "019b1000-0000-7000-8000-000000000106"
	retentionHeldDraft   = "019b1000-0000-7000-8000-000000000133"
	retentionReference   = "019b1000-0000-7000-8000-000000000107"
	retentionHold        = "019b1000-0000-7000-8000-000000000110"
)

func assertAnalyticsRetention(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	seedAnalyticsRetention(t, ctx, database)
	connection := analyticsRuntimeConnection(t, ctx, database)
	registerAnalyticsRetentionReference(t, ctx, connection)
	retireAnalyticsReferencedVersion(t, ctx, database)
	placeAnalyticsRetentionHold(t, ctx, connection)

	blocked := executeAnalyticsRetention(t, ctx, connection,
		"019b1000-0000-7000-8000-000000000111",
		"019b1000-0000-7000-8000-000000000112", 100)
	if !blocked.legalHoldBlocked || blocked.totalPurged() != 0 {
		t.Fatalf("blocked Analytics retention = %+v", blocked)
	}
	replayed := executeAnalyticsRetention(t, ctx, connection,
		"019b1000-0000-7000-8000-000000000111",
		"019b1000-0000-7000-8000-000000000113", 100)
	if replayed != blocked {
		t.Fatalf("Analytics retention replay changed: %#v != %#v", replayed, blocked)
	}
	assertAnalyticsRowCount(t, ctx, database,
		"t_analytics_a.analytics_saved_query_versions", "id", retentionOldDraft, 1)

	releaseAnalyticsRetentionHold(t, ctx, connection)
	completed := executeAnalyticsRetention(t, ctx, connection,
		"019b1000-0000-7000-8000-000000000114",
		"019b1000-0000-7000-8000-000000000115", 100)
	if completed.totalPurged() > 100 || completed.legalHoldBlocked {
		t.Fatalf("completed Analytics retention = %+v", completed)
	}
	assertAnalyticsRetentionBoundaries(t, ctx, database)
	assertAnalyticsRetentionGuards(t, ctx, database)

	releaseAnalyticsRetentionReference(t, ctx, connection)
	final := executeAnalyticsRetention(t, ctx, connection,
		"019b1000-0000-7000-8000-000000000116",
		"019b1000-0000-7000-8000-000000000117", 100)
	if final.versionsPurged != 1 {
		t.Fatalf("released Analytics reference purged %d versions, want 1", final.versionsPurged)
	}
	assertAnalyticsRowCount(t, ctx, database,
		"t_analytics_a.analytics_saved_query_versions", "id", retentionHeldVersion, 0)
	assertAnalyticsRetentionBounded(t, ctx, database, connection)
}
