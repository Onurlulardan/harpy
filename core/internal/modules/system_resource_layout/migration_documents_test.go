package systemresourcelayout

import (
	"strings"
	"testing"
)

func TestMigrationDocumentsEmbedFreshSQL(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	if len(documents) != 50 {
		t.Fatalf("migration document count = %d, want 50", len(documents))
	}
	seen := make(map[string]struct{}, len(documents))
	repairs := 0
	for _, document := range documents {
		path := document.Path()
		if document.Owner() != "system_resource_layout" || document.Scope() != "tenant" {
			t.Fatalf("owner/scope = %s/%s", document.Owner(), document.Scope())
		}
		if _, duplicate := seen[path]; duplicate {
			t.Fatalf("duplicate migration path %s", path)
		}
		if strings.Contains(path, "v1_adopted") ||
			!strings.HasSuffix(path, ".fresh.sql") ||
			len(document.Content()) == 0 {
			t.Fatalf("migration path = %s", path)
		}
		if strings.Contains(path, "20260730120104_") {
			repairs++
		}
		seen[path] = struct{}{}
	}
	if repairs != 4 {
		t.Fatalf("repair origin count = %d, want 4", repairs)
	}
}
