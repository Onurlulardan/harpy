package cutover

import (
	"path"
	"strings"
	"testing"
)

func TestGlobalBillingAuthorityMigrationPreservesOriginSemantics(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	sources := map[string]string{}
	for _, document := range documents {
		name := path.Base(document.Path())
		if !strings.HasPrefix(name, "20260716020103_") {
			continue
		}
		origin := "fresh"
		if strings.Contains(name, ".v1_adopted.sql") {
			origin = "v1_adopted"
		}
		sources[origin] = string(document.Content())
	}
	if len(sources) != 2 {
		t.Fatalf("global billing authority migration variants = %d, want 2", len(sources))
	}
	for origin, source := range sources {
		for _, table := range []string{
			"platform.domain_cutovers", "platform.cutover_groups",
			"platform.cutover_group_members",
		} {
			if !strings.Contains(source, "INSERT INTO "+table) {
				t.Fatalf("%s variant does not seed %s", origin, table)
			}
		}
		if strings.Contains(source, "ON CONFLICT") || strings.Contains(source, "-- +goose Down") {
			t.Fatalf("%s variant weakens immutable bootstrap semantics", origin)
		}
		if strings.Contains(source, "ELSE 1/0") ||
			!strings.Contains(source, "::pg_catalog.int4") {
			t.Fatalf("%s variant does not use a runtime cardinality assertion", origin)
		}
	}
	if !strings.Contains(sources["fresh"], "'v2', 'v2'") ||
		!strings.Contains(sources["fresh"], "acceptance_digest") ||
		!strings.Contains(sources["v1_adopted"], "'v1', 'v1'") {
		t.Fatal("origin-specific writer state differs")
	}
}
