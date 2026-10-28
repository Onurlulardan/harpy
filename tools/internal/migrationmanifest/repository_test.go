package migrationmanifest

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRepositoryMigrationManifestMatchesEveryOwnedSQLSource(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	inventory, err := LoadStorageInventory(root)
	if err != nil {
		t.Fatal(err)
	}
	groups, err := CollectAll(root, inventory)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := Build(inventory, groups)
	if err != nil {
		t.Fatal(err)
	}
	canonical, digest, err := Encode(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Entries) == 0 {
		t.Fatal("canonical migration manifest has no entries")
	}
	artifact := filepath.Join(
		root, "core", "internal", "platform", "migration", "release",
		digest+".migration-manifest.json",
	)
	published, err := os.ReadFile(artifact)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(published, canonical) {
		t.Fatal("published migration manifest differs from owned SQL sources")
	}
}
