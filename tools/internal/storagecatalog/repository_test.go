package storagecatalog

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestRepositoryStorageCatalogLoadsWithoutSemanticDrift(t *testing.T) {
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
	want := []string{
		"analytics", "billing", "catalog", "documents", "forms", "iam", "notifications",
		"platform_async", "platform_backupreceipt", "platform_cutover", "platform_migrator",
		"platform_postgres", "platformauthority", "platformops", "system_resource_layout",
		"tenancy", "work", "workflow",
	}
	if manifest.Version != catalogVersion || !reflect.DeepEqual(owners, want) {
		t.Fatalf("repository catalog version/owners = %d/%v", manifest.Version, owners)
	}
}
