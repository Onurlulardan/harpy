package catalog

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrationDocumentsPreserveOwnerScopeAndCount(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	if len(documents) != 106 {
		t.Fatalf("migration document count = %d, want 106", len(documents))
	}
	counts := map[string]int{}
	for _, document := range documents {
		if document.Owner() != "catalog" {
			t.Fatalf("owner = %q", document.Owner())
		}
		counts[document.Scope()]++
	}
	if counts["platform"] != 18 || counts["tenant"] != 88 || len(counts) != 2 {
		t.Fatalf("scope counts = %#v", counts)
	}
}

func TestMigrationTracksAreByteExact(t *testing.T) {
	for _, scope := range []string{"platform", "tenant"} {
		fresh, err := filepath.Glob(filepath.Join("migrations", scope, "fresh", "*.fresh.sql"))
		if err != nil || len(fresh) == 0 {
			t.Fatalf("%s fresh migrations: %v, count %d", scope, err, len(fresh))
		}
		for _, freshPath := range fresh {
			name := strings.Replace(filepath.Base(freshPath), ".fresh.sql", ".v1_adopted.sql", 1)
			adoptedPath := filepath.Join("migrations", scope, "v1_adopted", name)
			freshBody, readErr := os.ReadFile(freshPath)
			if readErr != nil {
				t.Fatal(readErr)
			}
			adoptedBody, readErr := os.ReadFile(adoptedPath)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if !bytes.Equal(freshBody, adoptedBody) {
				t.Errorf("migration tracks differ for %s", filepath.Base(freshPath))
			}
		}
	}
}

func TestTenantMigrationsDoNotReachOwnerTables(t *testing.T) {
	paths, err := filepath.Glob("migrations/tenant/fresh/*.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		body, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatal(readErr)
		}
		lower := strings.ToLower(string(body))
		for _, forbidden := range []string{
			"references iam_", "references forms_", "references workflow_", "references analytics_",
			"join iam_", "join forms_", "join workflow_", "join analytics_",
		} {
			if strings.Contains(lower, forbidden) {
				t.Errorf("%s contains owner-external SQL %q", filepath.Base(path), forbidden)
			}
		}
	}
}
