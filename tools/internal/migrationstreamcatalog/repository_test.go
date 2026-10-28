package migrationstreamcatalog

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestRepositoryMigrationStreamCatalogLoadsKnownOwners(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	owners := make([]string, len(manifest.Owners))
	for index, owner := range manifest.Owners {
		owners[index] = owner.Name
	}
	wantOwners := []string{
		"analytics", "billing", "catalog", "documents", "iam", "notifications",
		"platform_async", "platform_backupreceipt", "platform_cutover",
		"platform_migrator", "platformauthority", "platformops", "system_resource_layout",
		"tenancy", "work",
	}
	streams := manifest.Streams()
	if manifest.Version != catalogVersion || !reflect.DeepEqual(owners, wantOwners) {
		t.Fatalf("repository catalog version/owners = %d/%v", manifest.Version, owners)
	}
	if len(streams) == 0 {
		t.Fatal("repository catalog streams are empty")
	}
	if streams[0].VersionFrom != 20260713010000 {
		t.Fatalf("repository foundation stream moved: %#v", streams[0])
	}
}
