//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"sort"
	"strings"
	"testing"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

const asyncJobGuardLastVersion int64 = 20260715020910

func TestJobGuardMigrationsCompileOnPostgreSQL18(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx := context.Background()
			database := startAsyncMigrationDatabase(t, ctx, "job_guard_"+origin)
			prepareAsyncAuthorityFixture(t, ctx, database)
			addAsyncAuthorityStateColumns(t, ctx, database)
			for _, document := range jobGuardDocuments(t, origin) {
				if _, err := database.ExecContext(ctx, string(document.Content())); err != nil {
					t.Fatalf("apply %s: %v", document.Path(), err)
				}
			}
		})
	}
}

func addAsyncAuthorityStateColumns(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	const statement = `
ALTER TABLE platform.cutover_groups
    ADD COLUMN state text NOT NULL DEFAULT 'v2',
    ADD COLUMN writer text NOT NULL DEFAULT 'v2',
    ADD COLUMN generation bigint NOT NULL DEFAULT 1;
ALTER TABLE platform.domain_cutovers
    ADD COLUMN state text NOT NULL DEFAULT 'v2',
    ADD COLUMN writer text NOT NULL DEFAULT 'v2',
    ADD COLUMN generation bigint NOT NULL DEFAULT 1;`
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatalf("add async authority state fixture: %v", err)
	}
}

func jobGuardDocuments(t *testing.T, origin string) []migration.Document {
	t.Helper()
	documents, err := platformasync.MigrationDocuments()
	if err != nil {
		t.Fatalf("load async migrations: %v", err)
	}
	marker := "/migrations/" + origin + "/"
	selected := make([]migration.Document, 0, 13)
	for _, document := range documents {
		version := asyncDocumentVersion(t, document.Path())
		if strings.Contains(document.Path(), marker) &&
			version >= asyncSliceAFirstVersion && version <= asyncJobGuardLastVersion {
			selected = append(selected, document)
		}
	}
	sort.Slice(selected, func(left, right int) bool {
		return selected[left].Path() < selected[right].Path()
	})
	return selected
}
