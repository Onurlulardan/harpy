package tenancy

import (
	"path"
	"strings"
	"testing"
)

func TestTenantActivationV2MigrationIsPairedAndChainsCutoverV2(t *testing.T) {
	const version = "20260728103010"
	pair := map[string]string{}
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	for _, document := range documents {
		name := path.Base(document.Path())
		if !strings.HasPrefix(name, version+"_") {
			continue
		}
		origin := "fresh"
		if strings.Contains(name, ".v1_adopted.sql") {
			origin = "v1_adopted"
		}
		pair[origin] = string(document.Content())
	}
	if len(pair) != 2 || pair["fresh"] != pair["v1_adopted"] {
		t.Fatal("tenant activation v2 migration is not byte-identical and paired")
	}
	source := pair["fresh"]
	for _, marker := range []string{
		"CREATE FUNCTION platform.tenancy_activation_initialize_v2(",
		"platform.tenancy_activation_initialize_v1(",
		"platform.cutover_initialize_tenant_v2(",
		"GRANT EXECUTE ON FUNCTION platform.tenancy_activation_initialize_v2(",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("tenant activation v2 migration lacks %q", marker)
		}
	}
	if strings.Contains(source, "CREATE OR REPLACE") ||
		strings.Contains(source, "-- +goose Down") {
		t.Fatal("tenant activation v2 migration is not immutable and forward-only")
	}
}
