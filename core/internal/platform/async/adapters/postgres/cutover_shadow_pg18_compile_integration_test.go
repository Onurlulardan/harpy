//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"sort"
	"strings"
	"testing"
	"time"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

const (
	cutoverShadowLastVersion int64 = 20260718230709
)

func TestCutoverShadowMigrationsCompileOnPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			database := startAsyncMigrationDatabase(t, ctx, "shadow_"+origin)
			prepareCutoverShadowCompileFixture(t, ctx, database)
			applyCutoverShadowMigrations(t, ctx, database, origin)
			assertCutoverShadowCatalog(t, ctx, database)
			assertCutoverShadowBehavior(t, ctx, database)
		})
	}
}

func applyCutoverShadowMigrations(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	documents, err := platformasync.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	selected := cutoverShadowDocuments(t, documents, origin)
	for _, document := range selected {
		if _, err := database.ExecContext(ctx, string(document.Content())); err != nil {
			t.Fatalf("apply %s: %v", document.Path(), err)
		}
	}
}

func cutoverShadowDocuments(
	t *testing.T, documents []migration.Document, origin string,
) []migration.Document {
	t.Helper()
	marker := "/migrations/" + origin + "/"
	supersessionMarker := "/migrations/supersessions/" + origin + "/"
	selected := make([]migration.Document, 0, len(documents)/2)
	for _, document := range documents {
		version := asyncDocumentVersion(t, document.Path())
		if (strings.Contains(document.Path(), marker) ||
			strings.Contains(document.Path(), supersessionMarker)) &&
			version >= asyncSliceAFirstVersion && version <= cutoverShadowLastVersion {
			selected = append(selected, document)
		}
	}
	sort.Slice(selected, func(left, right int) bool {
		leftVersion := asyncDocumentVersion(t, selected[left].Path())
		rightVersion := asyncDocumentVersion(t, selected[right].Path())
		if leftVersion != rightVersion {
			return leftVersion < rightVersion
		}
		return selected[left].Path() < selected[right].Path()
	})
	return selected
}

func assertCutoverShadowCatalog(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	var functions, triggers int
	err := database.QueryRowContext(ctx, `SELECT
    count(*) FILTER (WHERE routine_name LIKE 'platform_async_%'),
    (SELECT count(*) FROM information_schema.triggers
     WHERE event_object_schema='platform'
       AND event_object_table='cutover_job_bindings')
FROM information_schema.routines WHERE routine_schema='platform'`).
		Scan(&functions, &triggers)
	if err != nil {
		t.Fatal(err)
	}
	if functions != 145 || triggers != 2 {
		t.Fatalf("cutover shadow catalog functions/triggers = %d/%d", functions, triggers)
	}
}
