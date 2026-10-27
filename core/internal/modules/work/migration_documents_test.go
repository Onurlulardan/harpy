package work

import (
	"strings"
	"testing"
)

func TestMigrationDocumentsEmbedFreshSQL(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	if len(documents) != 32 {
		t.Fatalf("migration document count = %d, want 32", len(documents))
	}
	seen := make(map[string]struct{}, len(documents))
	for _, document := range documents {
		path := document.Path()
		if document.Owner() != "work" || document.Scope() != "tenant" {
			t.Fatalf("owner/scope = %s/%s", document.Owner(), document.Scope())
		}
		if _, duplicate := seen[path]; duplicate {
			t.Fatalf("duplicate migration path %s", path)
		}
		if strings.Contains(path, "v1_adopted") ||
			!strings.HasSuffix(path, ".fresh.sql") ||
			!strings.Contains(path, "/migrations/tenant/fresh/") ||
			len(document.Content()) == 0 {
			t.Fatalf("migration path = %s", path)
		}
		seen[path] = struct{}{}
	}
}
