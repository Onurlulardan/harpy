package catalog

import (
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
	if len(documents) != 54 {
		t.Fatalf("migration document count = %d, want 54", len(documents))
	}
	counts := map[string]int{}
	for _, document := range documents {
		if document.Owner() != "catalog" {
			t.Fatalf("owner = %q", document.Owner())
		}
		if !strings.HasSuffix(document.Path(), ".fresh.sql") ||
			strings.Contains(document.Path(), "v1_adopted") {
			t.Fatalf("migration path %q is not fresh-only", document.Path())
		}
		counts[document.Scope()]++
	}
	if counts["platform"] != 9 || counts["tenant"] != 45 || len(counts) != 2 {
		t.Fatalf("scope counts = %#v", counts)
	}
}

func TestFreshMigrationTracksLoad(t *testing.T) {
	for _, scope := range []string{"platform", "tenant"} {
		fresh, err := filepath.Glob(filepath.Join("migrations", scope, "fresh", "*.fresh.sql"))
		if err != nil || len(fresh) == 0 {
			t.Fatalf("%s fresh migrations: %v, count %d", scope, err, len(fresh))
		}
		for _, freshPath := range fresh {
			freshBody, readErr := os.ReadFile(freshPath)
			if readErr != nil || len(freshBody) == 0 {
				t.Fatalf("fresh migration %s did not load: %v", freshPath, readErr)
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
