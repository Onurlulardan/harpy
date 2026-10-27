package iam

import "testing"

func TestCatalogAccessPermissionsMapOnlyExactOperations(t *testing.T) {
	for key, target := range catalogAccessTargets {
		if target.resource == "" || target.action == "" {
			t.Fatalf("Catalog access target %q is incomplete", key)
		}
	}
	if len(catalogAccessTargets) != 23 {
		t.Fatalf("Catalog access target count = %d", len(catalogAccessTargets))
	}
	for _, pair := range [][2]string{
		{"catalog.read", "catalog.installation.demo.get"},
		{"catalog.demo.manage", "catalog.installation.demo.removal_plan"},
		{"catalog.demo.manage", "catalog.installation.demo.install"},
		{"catalog.demo.manage", "catalog.installation.demo.remove"},
	} {
		if _, ok := resolveCatalogAccess(pair[0], pair[1]); !ok {
			t.Fatalf("Catalog demo pair %v is missing", pair)
		}
	}
	if _, ok := resolveCatalogAccess("catalog.read", "catalog.installation.install"); ok {
		t.Fatal("Catalog read permission authorized an install operation")
	}
	if _, ok := resolveCatalogAccess("catalog.read", "catalog.installation.demo.remove"); ok {
		t.Fatal("Catalog read permission authorized demo removal")
	}
	if !catalogOperationRequiresStepUp("catalog.installation.purge") ||
		!catalogOperationRequiresStepUp("catalog.installation.purge.approve") ||
		catalogOperationRequiresStepUp("catalog.installation.retire") {
		t.Fatal("Catalog step-up operation set differs from the contract")
	}
}
