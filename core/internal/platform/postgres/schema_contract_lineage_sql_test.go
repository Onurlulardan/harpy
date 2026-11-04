package postgres

// PRESERVE_NON_MIGRATION: asserts startup SQL does not read Goose native history.

import (
	"strings"
	"testing"
)

func TestSchemaContractProbeUsesBaselineIdentity(t *testing.T) {
	for _, removed := range []string{
		"platform.release_generations",
		"platform.schema_contracts",
		"platform.migration_runs",
		"platform.bootstrap_receipts",
	} {
		if strings.Contains(probeSchemaContractSQL, removed) {
			t.Fatalf("schema contract probe still reads %s", removed)
		}
	}
	if !strings.Contains(probeSchemaContractSQL, "platform.environment_identity AS identity") {
		t.Fatal("schema contract probe omitted environment identity")
	}
	if strings.Contains(probeSchemaContractSQL, "goose_db_version") ||
		strings.Contains(probeSchemaContractSQL, "20261028000000") {
		t.Fatal("schema contract probe still reads Goose history")
	}
}

func TestFoundationPreflightUsesBaselineIdentity(t *testing.T) {
	for _, removed := range []string{
		"platform.release_generations",
		"platform.schema_contracts",
		"platform.bootstrap_receipts",
	} {
		if strings.Contains(baselineFoundationSQL, removed) {
			t.Fatalf("foundation preflight still reads %s", removed)
		}
	}
	if strings.Contains(baselineFoundationSQL, "goose_db_version") ||
		strings.Contains(baselineFoundationSQL, "20261028000000") {
		t.Fatal("foundation preflight still reads Goose history")
	}
}
