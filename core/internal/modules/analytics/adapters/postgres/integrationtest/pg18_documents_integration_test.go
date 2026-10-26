//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"path"
	"sort"
	"strings"
	"testing"

	analyticsmodule "github.com/Onurlulardan/harpy/core/internal/modules/analytics"
	platformmigration "github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

func analyticsDocuments(
	t *testing.T, origin, scope string,
) []platformmigration.Document {
	t.Helper()
	documents, err := analyticsmodule.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	needle := "/" + origin + "/"
	selected := make([]platformmigration.Document, 0, len(documents)/2)
	for _, document := range documents {
		if document.Scope() == scope && strings.Contains(document.Path(), needle) {
			selected = append(selected, document)
		}
	}
	sort.Slice(selected, func(left, right int) bool {
		return selected[left].Path() < selected[right].Path()
	})
	if len(selected) == 0 {
		t.Fatalf("Analytics %s %s migration stream is empty", origin, scope)
	}
	return selected
}

func applyAnalyticsDocuments(
	t *testing.T, ctx context.Context, database *sql.DB, schema string,
	documents []platformmigration.Document,
) {
	t.Helper()
	for _, document := range documents {
		transaction, err := database.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		if schema != "" {
			if _, err = transaction.ExecContext(ctx,
				`SET LOCAL search_path = `+schema+`,platform,pg_catalog`); err != nil {
				_ = transaction.Rollback()
				t.Fatal(err)
			}
		}
		if _, err = transaction.ExecContext(ctx, string(document.Content())); err != nil {
			_ = transaction.Rollback()
			t.Fatalf("apply %s: %v", path.Base(document.Path()), err)
		}
		if err = transaction.Commit(); err != nil {
			t.Fatalf("commit %s: %v", path.Base(document.Path()), err)
		}
	}
}
