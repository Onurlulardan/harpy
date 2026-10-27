//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func assertHistoryTableAndAnchor(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	store *Store,
) {
	t.Helper()
	exists, err := store.TableExists(ctx, database)
	if err != nil {
		t.Fatalf("check history table: %v", err)
	}
	if !exists {
		t.Fatal("history table does not exist")
	}
	const query = `SELECT count(*) FROM platform.platform_migration_history
		WHERE version = 0 AND record_kind = 'anchor' AND is_applied
		AND owner IS NULL AND scope IS NULL
		AND origin IS NULL AND selected_path IS NULL
		AND selected_sha256 IS NULL AND run_id IS NULL
		AND target_database_generation IS NULL
		AND next_minimum_runtime_generation IS NULL`
	var count int
	if err := database.QueryRowContext(ctx, query).Scan(&count); err != nil {
		t.Fatalf("read history anchor: %v", err)
	}
	if count != 1 {
		t.Fatalf("valid version-zero anchor count = %d, want 1", count)
	}
}

func assertStoredMigrationMetadata(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	selected []postgresv1.SelectedMigration,
) {
	t.Helper()
	const query = `SELECT record_kind, is_applied, owner, scope, origin, selected_path,
		selected_sha256, run_id::text, target_database_generation,
		next_minimum_runtime_generation
		FROM platform.platform_migration_history WHERE version = $1`
	for _, expected := range selected {
		var recordKind, owner, scope, origin, path, digest, runID string
		var isApplied bool
		var targetGeneration, minimumRuntimeGeneration int64
		err := database.QueryRowContext(ctx, query, expected.Version).Scan(
			&recordKind, &isApplied, &owner, &scope, &origin, &path, &digest, &runID,
			&targetGeneration, &minimumRuntimeGeneration,
		)
		if err != nil {
			t.Fatalf("read migration %d metadata: %v", expected.Version, err)
		}
		actual := []any{recordKind, isApplied, owner, scope, origin, path, digest, runID,
			targetGeneration, minimumRuntimeGeneration}
		want := []any{"migration", true, expected.Owner, expected.Scope,
			string(expected.Origin),
			expected.Path, expected.SHA256, testRunID,
			expected.TargetDatabaseGeneration,
			expected.NextMinimumRuntimeGeneration}
		for index := range want {
			if actual[index] != want[index] {
				t.Fatalf("migration %d metadata[%d] = %v, want %v",
					expected.Version, index, actual[index], want[index])
			}
		}
	}
}
