package async

import (
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/internal/attestation/migrationsupersession"
)

func TestMigrationDocumentsExposeFreshSchema(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	if len(documents) == 0 {
		t.Fatal("fresh migration documents are empty")
	}
	const root = "core/internal/platform/async/migrations/"
	selected := make(map[string]bool, len(documents))
	for _, document := range documents {
		if document.Owner() != "platform_async" || document.Scope() != "platform" ||
			!strings.HasPrefix(document.Path(), root) ||
			strings.Contains(document.Path(), "/v1_adopted/") ||
			!strings.HasSuffix(document.Path(), ".fresh.sql") {
			t.Fatalf("unexpected migration document: %s %s", document.Owner(), document.Path())
		}
		selected[document.Path()] = true
	}
	catalog, err := migrationsupersession.DecodeCatalog(migrationSupersessionCatalog)
	if err != nil {
		t.Fatal(err)
	}
	freshEntries := 0
	for _, entry := range catalog.Entries() {
		if !strings.HasSuffix(entry.OldPath(), ".fresh.sql") {
			continue
		}
		freshEntries++
		if selected[entry.OldPath()] || !selected[entry.ReplacementPath()] {
			t.Fatalf("supersession selection differs: %s", entry.OldPath())
		}
	}
	if freshEntries == 0 {
		t.Fatal("fresh migration supersessions are missing")
	}
}
