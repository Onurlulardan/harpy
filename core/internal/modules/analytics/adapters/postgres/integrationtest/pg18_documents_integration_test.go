//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"path"
	"sort"
	"strings"
	"testing"
)

func analyticsDocuments(
	t *testing.T, origin, scope string,
) []appliedSQL {
	t.Helper()
	if origin != "fresh" && origin != "v1_adopted" {
		t.Fatalf("Analytics migration origin %q is unsupported", origin)
	}
	documents, err := []appliedSQL(nil), error(nil)
	if err != nil {
		t.Fatal(err)
	}
	needle := "/fresh/"
	selected := make([]appliedSQL, 0, len(documents)/2)
	for _, document := range documents {
		if document.Scope() == scope && strings.Contains(document.Path(), needle) {
			selected = append(selected, document)
		}
	}
	sort.Slice(selected, func(left, right int) bool {
		return selected[left].Path() < selected[right].Path()
	})
	if len(selected) == 0 {
		t.Fatalf("Analytics fresh %s migration stream is empty", scope)
	}
	return selected
}

func applyAnalyticsDocuments(
	t *testing.T, ctx context.Context, database *sql.DB, schema string,
	documents []appliedSQL,
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
