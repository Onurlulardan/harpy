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
	if err != nil || len(fresh) == 0 {
		t.Fatalf("Catalog function public ACL repair did not load: %v", err)
	}
	want := []byte("REVOKE ALL ON FUNCTION catalog_guard_purge_approval_v2() FROM PUBLIC;")
	if !bytes.Contains(fresh, want) {
		t.Fatal("Catalog function public ACL repair drifted")
	}
}
