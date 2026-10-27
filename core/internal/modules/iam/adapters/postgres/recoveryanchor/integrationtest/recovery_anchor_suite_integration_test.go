//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestRecoveryAnchorMigrationPairAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
			defer cancel()
			database := startRecoveryAnchorPG18(t, ctx)
			applyRecoveryAnchorMigrations(t, ctx, database, origin)
			assertRecoveryAnchorCatalog(t, ctx, database)
		})
	}
}

func assertRecoveryAnchorCatalog(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	var tables, functions, triggers int
	err := database.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM pg_catalog.pg_class
		 WHERE relnamespace='platform'::regnamespace
		   AND relname IN ('iam_recovery_anchors','iam_recovery_anchor_operations')),
		(SELECT COUNT(*) FROM pg_catalog.pg_proc
		 WHERE pronamespace='platform'::regnamespace
		   AND proname LIKE 'iam_recovery_anchor%_v1'),
		(SELECT COUNT(*) FROM pg_catalog.pg_trigger
		 WHERE tgname IN ('iam_recovery_anchors_guard','iam_recovery_anchor_operations_guard'))`).
		Scan(&tables, &functions, &triggers)
	if err != nil {
		t.Fatal(err)
	}
	if tables != 2 || functions != 11 || triggers != 2 {
		t.Fatalf("recovery anchor catalog = tables:%d functions:%d triggers:%d", tables, functions, triggers)
	}
}
