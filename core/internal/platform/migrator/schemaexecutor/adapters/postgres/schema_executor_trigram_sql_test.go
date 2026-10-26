package postgres

import (
	"strings"
	"testing"
)

func TestResolveTrigramSchemaAcceptsLiveCatalogWhenContractRunDiverges(t *testing.T) {
	sql := schemaExecutorResolveTrigramSchemaSQL
	for _, fragment := range []string{
		"platform.schema_contracts",
		"platform.extension_registry",
		"pg_catalog.pg_extension",
		"trigram_search",
		"LIMIT 1",
		"registry.verified_run_id = contract.verified_run_id",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("trigram registry SQL omitted %s", fragment)
		}
	}
	if strings.Contains(sql, "registry.verified_run_id = contract.verified_run_id\nJOIN") {
		t.Fatal("trigram registry SQL still requires matching verified_run_id before catalog join")
	}
}
