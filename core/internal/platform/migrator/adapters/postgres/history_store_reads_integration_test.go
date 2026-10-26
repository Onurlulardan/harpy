//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func assertHistoryReadSurface(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	store *Store,
	selected []postgresv1.SelectedMigration,
) {
	t.Helper()
	for _, expected := range selected {
		migration, err := store.GetMigration(ctx, database, expected.Version)
		if err != nil {
			t.Fatalf("get migration %d: %v", expected.Version, err)
		}
		if !migration.IsApplied || migration.Timestamp.IsZero() {
			t.Fatalf("migration %d read result is not applied with timestamp",
				expected.Version)
		}
	}
	migrations, err := store.ListMigrations(ctx, database)
	if err != nil {
		t.Fatalf("list migrations: %v", err)
	}
	wantVersions := []int64{selected[1].Version, selected[0].Version, 0}
	if len(migrations) != len(wantVersions) {
		t.Fatalf("listed migration count = %d, want %d",
			len(migrations), len(wantVersions))
	}
	for index, expected := range wantVersions {
		if migrations[index].Version != expected || !migrations[index].IsApplied {
			t.Fatalf("listed migration[%d] = (%d, %t), want (%d, true)",
				index, migrations[index].Version, migrations[index].IsApplied, expected)
		}
	}
	latest, err := store.GetLatestVersion(ctx, database)
	if err != nil {
		t.Fatalf("get latest migration: %v", err)
	}
	if latest != selected[1].Version {
		t.Fatalf("latest migration = %d, want %d", latest, selected[1].Version)
	}
}
