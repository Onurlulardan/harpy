//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
	"testing/fstest"

	"github.com/pressly/goose/v3"
)

func applySchemaExecutorMigrations(t *testing.T, ctx context.Context, database *sql.DB) {
	applySchemaExecutorMigrationsVariant(t, ctx, database, "fresh")
}

func applySchemaExecutorMigrationsVariant(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	variant string,
) {
	t.Helper()
	if _, err := database.ExecContext(ctx,
		`CREATE SCHEMA platform AUTHORIZATION harpy_owner`); err != nil {
		t.Fatalf("create platform schema for integration migrations: %v", err)
	}
	files := schemaExecutorMigrationFS(t, variant)
	goose.SetBaseFS(files)
	t.Cleanup(func() { goose.SetBaseFS(nil) })
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatal(err)
	}
	if err := goose.UpContext(ctx, database, "."); err != nil {
		t.Fatalf("apply schema executor integration migrations: %v", err)
	}
}

func schemaExecutorMigrationFS(t *testing.T, variant string) fstest.MapFS {
	t.Helper()
	if variant != "fresh" && variant != "v1_adopted" {
		t.Fatalf("unknown schema executor migration variant %q", variant)
	}
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve schema executor integration source")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(filename), "../../../../../../.."))
	directories := []string{
		"core/internal/platform/migrator/migrations/" + variant,
		"core/internal/platform/cutover/migrations/fresh",
		"core/internal/modules/tenancy/migrations/platform/fresh",
	}
	files := fstest.MapFS{}
	var names []string
	for _, directory := range directories {
		entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(directory)))
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" {
				names = append(names, filepath.Join(directory, entry.Name()))
			}
		}
	}
	sort.Strings(names)
	for _, name := range names {
		content, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		files[filepath.Base(name)] = &fstest.MapFile{Data: content, Mode: 0o444}
	}
	return files
}
