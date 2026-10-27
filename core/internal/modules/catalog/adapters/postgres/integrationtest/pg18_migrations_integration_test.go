//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	catalogmodule "github.com/Onurlulardan/harpy/core/internal/modules/catalog"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

func applyCatalogRuntimeFoundation(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	if origin != "fresh" && origin != "v1_adopted" {
		t.Fatalf("Catalog runtime foundation origin %q is unsupported", origin)
	}
	paths := []string{
		filepath.Join("..", "..", "..", "..", "tenancy", "migrations", "platform",
			"fresh", "20260718050600_create_tenancy_runtime_resolver_v1.fresh.sql"),
		filepath.Join("..", "..", "..", "..", "..", "platform", "cutover", "migrations",
			"fresh", "20260718050601_create_cutover_runtime_locks_v1.fresh.sql"),
	}
	for _, path := range paths {
		document, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = database.ExecContext(ctx, string(document)); err != nil {
			t.Fatalf("apply runtime foundation %s: %v", path, err)
		}
	}
}

func applyCatalogMigrations(
	t *testing.T, ctx context.Context, database *sql.DB, origin string,
) {
	t.Helper()
	documents, err := catalogmodule.MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(documents, func(left, right int) bool {
		return documents[left].Path() < documents[right].Path()
	})
	applyCatalogScopeDocuments(t, ctx, database, documents, "platform", origin, "")
	for _, tenant := range catalogTenants {
		applyCatalogScopeDocuments(t, ctx, database, documents, "tenant", origin, tenant.schema)
	}
}

func applyCatalogScopeDocuments(t *testing.T, ctx context.Context, database *sql.DB,
	documents []migration.Document, scope, origin, schema string,
) {
	t.Helper()
	if origin != "fresh" && origin != "v1_adopted" {
		t.Fatalf("Catalog migration origin %q is unsupported", origin)
	}
	needle := "/migrations/" + scope + "/fresh/"
	count := 0
	for _, document := range documents {
		if !strings.Contains(filepath.ToSlash(document.Path()), needle) {
			continue
		}
		applyCatalogMigration(t, ctx, database, schema, document)
		count++
	}
	want := 8
	if scope == "tenant" {
		want = 43
	}
	if count != want {
		t.Fatalf("Catalog %s/%s migration count=%d, want %d", scope, origin, count, want)
	}
}

func applyCatalogMigration(t *testing.T, ctx context.Context, database *sql.DB,
	schema string, document migration.Document,
) {
	t.Helper()
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = transaction.Rollback() }()
	if schema != "" {
		_, err = transaction.ExecContext(ctx,
			"SELECT set_config('search_path',$1,true)", schema+",platform,pg_catalog")
	}
	if err == nil {
		_, err = transaction.ExecContext(ctx, string(document.Content()))
	}
	if err != nil {
		t.Fatalf("apply %s to %s: %v", document.Path(), schema, err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}
