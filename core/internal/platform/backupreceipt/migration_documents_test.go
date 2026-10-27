package backupreceipt

import (
	"strings"
	"testing"
)

func TestMigrationDocumentsExposeFreshSchema(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	if len(documents) == 0 {
		t.Fatal("fresh migration documents are empty")
	}
	const root = "core/internal/platform/backupreceipt/migrations/fresh/"
	for _, document := range documents {
		if document.Owner() != "platform_backupreceipt" || document.Scope() != "platform" ||
			!strings.HasPrefix(document.Path(), root) ||
			!strings.HasSuffix(document.Path(), ".fresh.sql") {
			t.Fatalf("unexpected migration document: %s %s", document.Owner(), document.Path())
		}
	}
}
