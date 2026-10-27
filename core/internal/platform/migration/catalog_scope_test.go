package migration

import (
	"io/fs"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestPartitionCatalogNeverMixesPlatformAndTenantMigrations(t *testing.T) {
	platformPath := "owner/migrations/fresh/20260713010000_platform.fresh.sql"
	tenantPath := "owner/migrations/fresh/20260713050000_tenant.fresh.sql"
	bound := map[string][]byte{
		platformPath: testFreshSQL,
		tenantPath:   testFreshSQL,
	}
	files, selected, err := partitionCatalog(bound, []postgresv1.SelectedMigration{
		{Scope: "platform", Path: platformPath},
		{Scope: "tenant", Path: tenantPath},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(selected["platform"]) != 1 || len(selected["tenant"]) != 1 {
		t.Fatalf("selected migrations crossed scope: %#v", selected)
	}
	if _, err := fs.ReadFile(files["platform"], "20260713050000_tenant.fresh.sql"); err == nil {
		t.Fatal("tenant migration leaked into platform Goose filesystem")
	}
	if _, err := fs.ReadFile(files["tenant"], "20260713010000_platform.fresh.sql"); err == nil {
		t.Fatal("platform migration leaked into tenant Goose filesystem")
	}
}

func TestPartitionCatalogRejectsScopeAndSelectionDrift(t *testing.T) {
	path := "owner/migrations/fresh/20260713010000_platform.fresh.sql"
	tests := []struct {
		name     string
		bound    map[string][]byte
		selected []postgresv1.SelectedMigration
	}{
		{"invalid scope", map[string][]byte{path: testFreshSQL},
			[]postgresv1.SelectedMigration{{Scope: "shared", Path: path}}},
		{"unselected source", map[string][]byte{path: testFreshSQL}, nil},
		{"duplicate selection", map[string][]byte{path: testFreshSQL},
			[]postgresv1.SelectedMigration{
				{Scope: "platform", Path: path}, {Scope: "platform", Path: path},
			}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, _, err := partitionCatalog(test.bound, test.selected); err == nil {
				t.Fatal("migration scope or selection drift was accepted")
			}
		})
	}
}
