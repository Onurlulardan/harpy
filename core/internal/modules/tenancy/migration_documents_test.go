package tenancy

import (
	"path"
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
		if document.Owner() != "tenancy" {
			t.Fatalf("owner = %q", document.Owner())
		}
		if !strings.HasSuffix(document.Path(), ".fresh.sql") ||
			strings.Contains(document.Path(), "v1_adopted") {
			t.Fatalf("migration path %q is not fresh-only", document.Path())
		}
		counts[document.Scope()]++
	}
	if len(documents) != 133 || counts["platform"] != 129 || counts["tenant"] != 4 ||
		len(counts) != 2 {
		t.Fatalf("migration documents = %d, scopes %#v", len(documents), counts)
	}
}

func freshMigrationDocument(t *testing.T, versionPrefix string) string {
	t.Helper()
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	var source string
	matches := 0
	for _, document := range documents {
		name := path.Base(document.Path())
		if !strings.HasPrefix(name, versionPrefix+"_") {
			continue
		}
		if !strings.HasSuffix(name, ".fresh.sql") ||
			strings.Contains(document.Path(), "v1_adopted") {
			t.Fatalf("migration %s is not fresh-only", document.Path())
		}
		if len(document.Content()) == 0 {
			t.Fatalf("fresh migration %s is empty", document.Path())
		}
		matches++
		source = string(document.Content())
	}
	if matches != 1 {
		t.Fatalf("fresh migration %s count = %d", versionPrefix, matches)
	}
	return source
}
