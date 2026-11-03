package postgres

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
	for _, required := range []string{
		"platform.environment_identity AS identity",
		"platform.goose_db_version AS history",
		"20261028000000, 20261102000000",
	} {
		if !strings.Contains(probeSchemaContractSQL, required) {
			t.Fatalf("schema contract probe omitted %q", required)
		}
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
	if !strings.Contains(baselineFoundationSQL, "20261028000000, 20261102000000") {
		t.Fatal("foundation preflight must require the platform baseline versions")
	}
}
