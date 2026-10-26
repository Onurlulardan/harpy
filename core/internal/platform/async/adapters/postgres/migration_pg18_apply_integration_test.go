//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"sort"
	"strconv"
	"strings"
	"testing"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

const (
	asyncSliceAFirstVersion int64 = 20260715020000
	asyncSliceALastVersion  int64 = 20260715020600
)

func applyAsyncSliceA(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	origin string,
) {
	t.Helper()
	prepareAsyncAuthorityFixture(t, ctx, database)
	documents, err := platformasync.MigrationDocuments()
	if err != nil {
		t.Fatalf("load async migration documents: %v", err)
	}
	selected := selectAsyncSliceA(t, documents, origin)
	for _, document := range selected {
		transaction, err := database.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin %s: %v", document.Path(), err)
		}
		if _, err := transaction.ExecContext(ctx, string(document.Content())); err != nil {
			_ = transaction.Rollback()
			t.Fatalf("apply %s: %v", document.Path(), err)
		}
		if err := transaction.Commit(); err != nil {
			t.Fatalf("commit %s: %v", document.Path(), err)
		}
	}
}

func selectAsyncSliceA(
	t *testing.T,
	documents []migration.Document,
	origin string,
) []migration.Document {
	t.Helper()
	if origin != "fresh" && origin != "v1_adopted" {
		t.Fatalf("unsupported async migration origin: %s", origin)
	}
	selected := make([]migration.Document, 0, 7)
	marker := "/migrations/" + origin + "/"
	for _, document := range documents {
		version := asyncDocumentVersion(t, document.Path())
		if strings.Contains(document.Path(), marker) &&
			version >= asyncSliceAFirstVersion && version <= asyncSliceALastVersion {
			selected = append(selected, document)
		}
	}
	sort.Slice(selected, func(left, right int) bool {
		return selected[left].Path() < selected[right].Path()
	})
	if len(selected) != 7 {
		t.Fatalf("%s Slice A migration count = %d, want 7", origin, len(selected))
	}
	return selected
}

func asyncDocumentVersion(t *testing.T, path string) int64 {
	t.Helper()
	base := path[strings.LastIndex(path, "/")+1:]
	if len(base) < 14 {
		t.Fatalf("async migration path has no version: %s", path)
	}
	version, err := strconv.ParseInt(base[:14], 10, 64)
	if err != nil {
		t.Fatalf("parse async migration version %s: %v", path, err)
	}
	return version
}
