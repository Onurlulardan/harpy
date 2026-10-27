package migration

import (
	"os"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/pressly/goose/v3"
)

func TestCatalogFilesystemIsDiscoverableByGooseProvider(t *testing.T) {
	bound := map[string][]byte{
		"core/internal/platform/migrator/migrations/fresh/" +
			"20260713010000_create_foundation.fresh.sql": testFreshSQL,
	}
	files, selected, err := partitionCatalog(bound, []postgresv1.SelectedMigration{{
		Scope: "platform",
		Path: "core/internal/platform/migrator/migrations/fresh/" +
			"20260713010000_create_foundation.fresh.sql",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(selected["platform"]) != 1 {
		t.Fatal("selected platform migration was not isolated by scope")
	}
	goose.SetBaseFS(files["platform"])
	t.Cleanup(func() { goose.SetBaseFS(os.DirFS(".")) })
	migrations, err := goose.CollectMigrations(".", 0, goose.MaxVersion)
	if err != nil || len(migrations) != 1 {
		t.Fatalf("Goose could not discover catalog migration: %v", err)
	}
}

func TestFlattenDocumentsRejectsDuplicateBasename(t *testing.T) {
	if _, err := flattenDocuments(map[string][]byte{
		"first/20260713010000_same.fresh.sql":  testFreshSQL,
		"second/20260713010000_same.fresh.sql": testFreshSQL,
	}); err == nil {
		t.Fatal("duplicate Goose migration basename was accepted")
	}
}
