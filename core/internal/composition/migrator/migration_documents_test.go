package migrator

import (
	"strings"
	"testing"
)

func TestDeployMigrationDocumentsSelectFreshSchema(t *testing.T) {
	documents, err := deployMigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	if len(documents) == 0 {
		t.Fatal("fresh migration documents are empty")
	}
	for _, document := range documents {
		if strings.Contains(document.Path(), "/v1_adopted/") ||
			strings.HasSuffix(document.Path(), ".v1_adopted.sql") ||
			strings.Contains(document.Path(), "/modules/forms/") ||
			strings.Contains(document.Path(), "/modules/workflow/") ||
			strings.Contains(document.Path(), "/modules/ai/") ||
			strings.Contains(document.Path(), "/modules/mcp/") {
			t.Fatalf("fresh migration path is outside the startup set: %s", document.Path())
		}
	}
}
