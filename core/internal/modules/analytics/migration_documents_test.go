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
	if len(documents) != 204 || counts["platform"] != 26 || counts["tenant"] != 178 ||
		len(counts) != 2 {
		t.Fatalf("migration documents = %d, scopes %#v", len(documents), counts)
	}
}

func analyticsMigrationIdentity(t *testing.T, scope, path string) (string, string) {
	t.Helper()
	root := "core/internal/modules/analytics/migrations/" + scope + "/"
	parts := strings.Split(strings.TrimPrefix(path, root), "/")
	if len(parts) == 3 && parts[0] == "supersessions" {
		parts = parts[1:]
	}
	if len(parts) != 2 {
		t.Fatalf("migration path is not canonical: %s", path)
	}
	origin := parts[0]
	suffix := "." + origin + ".sql"
	if (origin != "fresh" && origin != "v1_adopted") ||
		!strings.HasSuffix(parts[1], suffix) {
		t.Fatalf("migration origin is invalid: %s", path)
	}
	return origin, strings.TrimSuffix(parts[1], suffix)
}
