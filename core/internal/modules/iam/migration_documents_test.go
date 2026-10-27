package iam

import (
	"strings"
	"testing"
)

func TestMigrationDocumentsEmbedFreshSQL(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	if len(documents) != 729 {
		t.Fatalf("migration document count = %d, want 729", len(documents))
	}
	counts := map[string]int{}
	seen := map[string]struct{}{}
	for _, document := range documents {
		path := document.Path()
		if document.Owner() != "iam" || seenPath(seen, path) ||
			!freshMigrationPath(document.Scope(), path) ||
			len(document.Content()) == 0 {
			t.Fatalf("unexpected migration %s %s %s",
				document.Owner(), document.Scope(), path)
		}
		seen[path] = struct{}{}
		counts[document.Scope()]++
	}
	if counts["platform"] != 103 || counts["tenant"] != 626 {
		t.Fatalf("scope counts = platform %d tenant %d",
			counts["platform"], counts["tenant"])
	}
}

func seenPath(seen map[string]struct{}, path string) bool {
	_, exists := seen[path]
	return exists
}

func freshMigrationPath(scope, path string) bool {
	root := "core/internal/modules/iam/migrations/" + scope + "/"
	return (scope == "platform" || scope == "tenant") &&
		strings.HasPrefix(path, root) &&
		strings.Contains(path, "/fresh/") &&
		strings.HasSuffix(path, ".fresh.sql") &&
		!strings.Contains(path, "v1_adopted")
}
