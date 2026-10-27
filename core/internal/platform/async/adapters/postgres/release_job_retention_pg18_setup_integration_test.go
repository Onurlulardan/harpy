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

const retentionReleaseLastVersion int64 = 20260715021250

func TestJobRetentionReleaseAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx := context.Background()
			database := startAsyncMigrationDatabase(t, ctx, "retention_release_"+origin)
			prepareAsyncAuthorityFixture(t, ctx, database)
			addAsyncAuthorityStateColumns(t, ctx, database)
			applyRetentionReleaseMigrations(t, ctx, database, origin)
			seedRetentionReleaseFixture(t, ctx, database)
			grantRetentionReleaseRuntime(t, ctx, database)
			connection, err := database.Conn(ctx)
			if err != nil {
				t.Fatalf("open retention release connection: %v", err)
			}
			defer connection.Close()
			if _, err := connection.ExecContext(
				ctx, "SET SESSION AUTHORIZATION harpy_v2_runtime",
			); err != nil {
				t.Fatalf("activate runtime session: %v", err)
			}
			assertRetentionReleaseCases(t, ctx, connection)
		})
	}
}

func applyRetentionReleaseMigrations(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	documents, err := platformasync.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	selected := make([]migration.Document, 0, 24)
	marker := "/migrations/" + origin + "/"
	for _, document := range documents {
		version := asyncDocumentVersion(t, document.Path())
		if strings.Contains(document.Path(), marker) &&
			version >= asyncSliceAFirstVersion && version <= retentionReleaseLastVersion {
			selected = append(selected, document)
		}
	}
	sort.Slice(selected, func(left, right int) bool {
		return selected[left].Path() < selected[right].Path()
	})
	for _, document := range selected {
		if _, err := database.ExecContext(ctx, string(document.Content())); err != nil {
			t.Fatalf("apply %s: %v", document.Path(), err)
		}
	}
}

func grantRetentionReleaseRuntime(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	const statement = `
CREATE ROLE harpy_v2_runtime;
GRANT USAGE ON SCHEMA platform TO harpy_v2_runtime;
GRANT SELECT ON TABLE platform.jobs TO harpy_v2_runtime;
GRANT EXECUTE ON FUNCTION platform.platform_async_release_job_retention_v1(
    uuid,uuid,text,bigint,text,bigint,bigint,text,timestamptz
) TO harpy_v2_runtime;`
	if _, err := database.ExecContext(ctx, statement); err != nil {
		t.Fatalf("grant retention release runtime: %v", err)
	}
}
