package storagecatalog

import (
	"testing"
)

func repositoryOwner(t *testing.T, manifest Manifest, name string) Owner {
	t.Helper()
	for _, owner := range manifest.Owners {
		if owner.Name == name {
			return owner
		}
	}
	t.Fatalf("repository owner %q is missing", name)
	return Owner{}
}

func assertCatalogBoundary(
	t *testing.T, label, first, wantFirst, last, wantLast string,
) {
	t.Helper()
	if first != wantFirst || last != wantLast {
		t.Fatalf("%s boundary = %q/%q", label, first, last)
	}
}

func requireCatalogCount(t *testing.T, label string, count, want int) {
	t.Helper()
	if count != want {
		t.Fatalf("%s count = %d, want %d", label, count, want)
	}
}
