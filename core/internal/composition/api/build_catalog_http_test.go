package api

import "testing"

func TestBuildCatalogOwnersClosesExactOwnerDirectory(t *testing.T) {
	directory, err := buildCatalogOwners()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"analytics", "catalog", "iam"} {
		owner, resolveErr := directory.Resolve(name)
		if resolveErr != nil || owner == nil {
			t.Fatalf("Catalog owner %s unavailable: %v", name, resolveErr)
		}
	}
	if _, err := directory.Resolve("billing"); err == nil {
		t.Fatal("unsupported Catalog owner resolved")
	}
}
