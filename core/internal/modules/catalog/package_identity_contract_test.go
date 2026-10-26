package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCatalogHTTPPackageIdentityContractIsStableAndNonZero(t *testing.T) {
	root := catalogRepositoryRoot(t)
	schema, err := os.ReadFile(filepath.Join(
		root, "contracts", "http", "catalog-v2.schemas.json",
	))
	if err != nil {
		t.Fatal(err)
	}
	document := string(schema)
	if strings.Count(document, `#/$defs/StableUUID`) != 9 {
		t.Fatalf("StableUUID reference count = %d, want 9",
			strings.Count(document, `#/$defs/StableUUID`))
	}
	for _, required := range []string{
		`"StableUUID":{"type":"string","pattern":"^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"`,
		`"not":{"const":"00000000-0000-0000-0000-000000000000"}`,
		`"PackageIDParameter":{"name":"package_id","in":"path","required":true,"schema":{"$ref":"#/$defs/StableUUID"}}`,
		`"PackageAfterParameter":{"name":"after","in":"query","required":false,"schema":{"$ref":"#/$defs/StableUUID"}}`,
	} {
		if !strings.Contains(document, required) {
			t.Fatalf("Catalog schema lacks %s", required)
		}
	}
	if strings.Contains(document, `"package_id":{"$ref":"#/$defs/UUIDv7"}`) {
		t.Fatal("Catalog package_id still references UUIDv7")
	}
	paths, err := os.ReadFile(filepath.Join(
		root, "contracts", "http", "catalog-v2.paths.json",
	))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(paths),
		`catalog-v2.schemas.json#/$defs/PackageAfterParameter`) {
		t.Fatal("Catalog package pagination still uses the UUIDv7 cursor")
	}
}
