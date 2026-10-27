package worker

import "testing"

func TestBuildCatalogWorkerOwnersClosesExactDirectory(t *testing.T) {
	directory, err := buildCatalogWorkerOwners()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"analytics", "catalog", "iam"} {
		owner, resolveErr := directory.Resolve(name)
		if resolveErr != nil || owner == nil {
			t.Fatalf("Catalog worker owner %s unavailable: %v", name, resolveErr)
		}
	}
}
