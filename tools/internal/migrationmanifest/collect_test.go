package migrationmanifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollectReadsOnlyRegularMigrationSources(t *testing.T) {
	root := t.TempDir()
	directory := "core/internal/platform/migrator/migrations"
	for origin, statement := range map[string]string{
		"fresh": "SELECT 1;", "v1_adopted": "SELECT 2;",
	} {
		path := filepath.Join(root, filepath.FromSlash(directory), origin)
		if err := os.MkdirAll(path, 0o700); err != nil {
			t.Fatal(err)
		}
		filename := filepath.Join(path, testVersion+"_create_foundation."+origin+".sql")
		if err := os.WriteFile(filename, []byte("-- +goose Up\n"+statement+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	group, err := Collect(root, SourceSpec{
		Owner: testOwner, Scope: "platform", Phase: "foundation",
		TargetDatabaseGeneration: 1, NextMinimumRuntimeGeneration: 1,
		VersionFrom: 20260713010000, VersionTo: 20260713010000,
		Directory: directory,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(group.Documents) != 2 || group.Documents[0].Path() >= group.Documents[1].Path() {
		t.Fatalf("collected migration sources are invalid: %#v", group.Documents)
	}
	if _, err := Build(testStorageInventory(group), []SourceGroup{group}); err != nil {
		t.Fatal(err)
	}
}

func TestCollectRejectsUnexpectedNonSQLFile(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "migrations")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "README.md"), []byte("unsafe"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Collect(root, SourceSpec{Owner: testOwner, Directory: "migrations"}); err == nil {
		t.Fatal("unexpected migration directory file was accepted")
	}
}
