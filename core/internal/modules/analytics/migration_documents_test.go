package analytics

import (
	"strings"
	"testing"
)

func TestMigrationDocumentsReturnFreshSQL(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, document := range documents {
		if document.Owner() != "analytics" {
			t.Fatalf("owner = %q", document.Owner())
		}
		if !strings.HasSuffix(document.Path(), ".fresh.sql") ||
			strings.Contains(document.Path(), "v1_adopted") {
			t.Fatalf("migration path %q is not fresh-only", document.Path())
		}
		counts[document.Scope()]++
	}
	if len(documents) != 197 || counts["platform"] != 26 || counts["tenant"] != 171 ||
		len(counts) != 2 {
		t.Fatalf("migration documents = %d, scopes %#v", len(documents), counts)
	}
}
