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
	var source string
	for _, document := range documents {
		name := path.Base(document.Path())
		if strings.HasPrefix(name, "20260716020103_") && strings.HasSuffix(name, ".fresh.sql") {
			source = string(document.Content())
		}
	}
	if source == "" {
		t.Fatal("fresh global billing authority migration did not load")
	}
	for _, table := range []string{
		"platform.domain_cutovers", "platform.cutover_groups",
		"platform.cutover_group_members",
	} {
		if !strings.Contains(source, "INSERT INTO "+table) {
			t.Fatalf("fresh migration does not seed %s", table)
		}
	}
	if strings.Contains(source, "ON CONFLICT") || strings.Contains(source, "-- +goose Down") {
		t.Fatal("fresh migration weakens immutable bootstrap semantics")
	}
	if strings.Contains(source, "ELSE 1/0") || !strings.Contains(source, "::pg_catalog.int4") {
		t.Fatal("fresh migration does not use a runtime cardinality assertion")
	}
	if !strings.Contains(source, "'v2', 'v2'") || !strings.Contains(source, "acceptance_digest") {
		t.Fatal("fresh writer state did not load")
	}
}
