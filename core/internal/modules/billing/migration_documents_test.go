package billing

import (
	"strings"
	"testing"
)

func TestMigrationDocumentsReturnFreshSQL(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	if len(documents) != 208 {
		t.Fatalf("migration document count = %d, want 208", len(documents))
	}
	for _, document := range documents {
		if document.Owner() != "billing" || document.Scope() != "platform" {
			t.Fatalf("owner/scope = %s/%s", document.Owner(), document.Scope())
		}
		if !strings.HasSuffix(document.Path(), ".fresh.sql") ||
			strings.Contains(document.Path(), "v1_adopted") {
			t.Fatalf("migration path %q is not fresh-only", document.Path())
		}
	}
}
