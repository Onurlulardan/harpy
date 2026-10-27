package catalog

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestCatalogFunctionPublicACLRepairIsExactAndPaired(t *testing.T) {
	const name = "20260802232201_harden_catalog_function_public_acl_v1"
	fresh, err := os.ReadFile(filepath.Join(
		"migrations", "tenant", "fresh", name+".fresh.sql",
	))
	if err != nil {
		t.Fatal(err)
	}
	adopted, err := os.ReadFile(filepath.Join(
		"migrations", "tenant", "v1_adopted", name+".v1_adopted.sql",
	))
	if err != nil {
		t.Fatal(err)
	}
	want := []byte("REVOKE ALL ON FUNCTION catalog_guard_purge_approval_v2() FROM PUBLIC;")
	if !bytes.Equal(fresh, adopted) || !bytes.Contains(fresh, want) {
		t.Fatal("Catalog function public ACL repair drifted")
	}
}
