package notifications

import (
	"strings"
	"testing"
)

func TestMigrationDocumentsEmbedFreshSQL(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	if len(documents) != 34 {
		t.Fatalf("migration document count = %d, want 34", len(documents))
	}
	counts := map[string]int{}
	seen := make(map[string]struct{}, len(documents))
	for _, document := range documents {
		if document.Owner() != "notifications" {
			t.Fatalf("owner = %s", document.Owner())
		}
		scope := document.Scope()
		path := document.Path()
		if _, duplicate := seen[path]; duplicate ||
			(scope != "platform" && scope != "tenant") ||
			strings.Contains(path, "v1_adopted") ||
			!strings.HasSuffix(path, ".fresh.sql") ||
			!strings.Contains(path, "/migrations/"+scope+"/fresh/") ||
			len(document.Content()) == 0 {
			t.Fatalf("migration path = %s scope = %s", path, scope)
		}
		seen[path] = struct{}{}
		counts[scope]++
	}
	if counts["platform"] != 3 || counts["tenant"] != 31 {
		t.Fatalf("scope counts = %v", counts)
	}
}
