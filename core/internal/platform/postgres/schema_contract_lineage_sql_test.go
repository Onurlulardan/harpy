package postgres

import (
	"strings"
	"testing"
)

func TestSchemaContractProbeRequiresActiveAndHistoricalReleaseEvidence(t *testing.T) {
	for _, required := range []string{
		"platform.release_generations AS active_release",
		"platform.release_generations AS compatible_release",
		"platform.migration_runs AS run_row",
		"run_row.command = 'up'",
		"run_row.state = 'succeeded'",
		"bootstrap_receipt_run_id",
		"extension_receipt_run_id",
		"cleanup_receipt_run_id",
		"compatible_release.runtime_generation = $1",
		"compatible_release.minimum_database_generation = $2",
		"compatible_release.postgres_image_digest = $5",
	} {
		if !strings.Contains(probeSchemaContractSQL, required) {
			t.Fatalf("schema contract probe omitted %q", required)
		}
	}
}

func TestFoundationPreflightDoesNotTreatBootstrapReleaseAsCurrent(t *testing.T) {
	if strings.Contains(preflightFoundationSQL, "run.release_id = identity.release_id") {
		t.Fatal("foundation preflight compares current run to bootstrap release")
	}
	for _, required := range []string{
		"active_release.release_id = run.release_id",
		"bootstrap_receipt.run_id = active_release.bootstrap_receipt_run_id",
		"compatible_release.migration_manifest_sha256 = $4",
	} {
		if !strings.Contains(preflightFoundationSQL, required) {
			t.Fatalf("foundation preflight omitted %q", required)
		}
	}
}
