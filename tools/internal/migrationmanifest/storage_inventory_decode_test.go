package migrationmanifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

type storageOwnerFixture struct {
	name      string
	directory string
	scope     string
}

func TestLoadStorageInventoryUsesV2Catalog(t *testing.T) {
	root := t.TempDir()
	owner := storageOwnerFixture{
		name: "alpha", directory: "core/internal/platform/alpha/migrations",
		scope: "platform",
	}
	writeStorageCatalogFixture(t, root, []storageOwnerFixture{owner})
	writeMigrationStreamsFixture(t, root, owner)

	inventory, err := LoadStorageInventory(root)
	if err != nil {
		t.Fatalf("load storage inventory: %v", err)
	}
	claims := inventory.Claims()
	if len(claims) != 1 || claims[0].Owner != owner.name ||
		claims[0].Directory != owner.directory || claims[0].Scope != owner.scope {
		t.Fatalf("unexpected storage claims: %#v", claims)
	}
	if len(inventory.Streams()) != 1 {
		t.Fatal("storage inventory omitted migration stream")
	}
}

func TestLoadStorageInventoryRejectsInvalidV2CatalogClaims(t *testing.T) {
	tests := []struct {
		name   string
		owners []storageOwnerFixture
		want   string
	}{
		{
			name: "duplicate directory across owners",
			owners: []storageOwnerFixture{
				{name: "alpha", directory: "migrations/shared", scope: "platform"},
				{name: "beta", directory: "migrations/shared", scope: "platform"},
			},
			want: "claimed more than once",
		},
		{
			name: "invalid directory",
			owners: []storageOwnerFixture{
				{name: "alpha", directory: "../migrations", scope: "platform"},
			},
			want: "is invalid",
		},
		{
			name: "invalid scope",
			owners: []storageOwnerFixture{
				{name: "alpha", directory: "migrations/alpha", scope: "global"},
			},
			want: "invalid scope",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			writeStorageCatalogFixture(t, root, test.owners)
			_, err := LoadStorageInventory(root)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected %q error, got %v", test.want, err)
			}
		})
	}
}

func writeStorageCatalogFixture(
	t *testing.T,
	root string,
	owners []storageOwnerFixture,
) {
	t.Helper()
	owners = append([]storageOwnerFixture(nil), owners...)
	sort.Slice(owners, func(left, right int) bool {
		return owners[left].name < owners[right].name
	})
	includes := make([]string, 0, len(owners))
	for _, owner := range owners {
		include := "contracts/storage/owners/" + owner.name + ".json"
		includes = append(includes, include)
		writeInventoryFixtureFile(t, root, include, ownerFragment(owner))
	}
	index, err := json.Marshal(struct {
		Version  int      `json:"version"`
		Includes []string `json:"includes"`
	}{Version: 2, Includes: includes})
	if err != nil {
		t.Fatal(err)
	}
	writeInventoryFixtureFile(
		t, root, "contracts/storage/owners.yaml", string(index),
	)
}

func ownerFragment(owner storageOwnerFixture) string {
	return fmt.Sprintf(`{
  "version": 1,
  "owner": {
    "name": %q,
    "prefix": %q,
    "scopes": [%q],
    "migration_dirs": [{"path": %q, "scope": %q}],
    "static_sql_dirs": [],
    "platform_schemas": [],
    "schema_usage_grants": [],
    "platform_routines": [],
    "routine_execute_grants": [],
    "platform_tables": [],
    "external_accesses": [],
    "foreign_keys": [],
    "legacy_tables": [],
    "dynamic_namespaces": []
  }
}`, owner.name, owner.name+"_", owner.scope, owner.directory, owner.scope)
}

func writeMigrationStreamsFixture(
	t *testing.T,
	root string,
	owner storageOwnerFixture,
) {
	t.Helper()
	include := "contracts/storage/migration-streams/" + owner.name + ".json"
	rootContent := fmt.Sprintf(
		`{"version":2,"includes":[%q]}`, include,
	)
	fragment := fmt.Sprintf(`{"version":1,"owner":%q,"streams":[{
    "owner":%q,"directory":%q,"scope":%q,
    "version_from":1,"version_to":1,"phase":"foundation",
    "target_database_generation":1,"next_minimum_runtime_generation":1
	  }]}`, owner.name, owner.name, owner.directory, owner.scope)
	writeInventoryFixtureFile(
		t, root, "contracts/storage/migration-streams.yaml", rootContent,
	)
	writeInventoryFixtureFile(t, root, include, fragment)
}

func writeInventoryFixtureFile(t *testing.T, root, name, content string) {
	t.Helper()
	filename := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(filename), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
