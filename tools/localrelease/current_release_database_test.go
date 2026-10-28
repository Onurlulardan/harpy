package main

import (
	"strings"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func TestCurrentReleaseStatusRejectsFoundationContract(t *testing.T) {
	contract := currentReleaseStatusTestContract()
	contract.MigrationManifestSHA256 = foundationMigrationDigest
	if err := validateCurrentReleaseContract(contract); err == nil {
		t.Fatal("foundation contract was accepted as the current release")
	}
}

func TestCurrentReleaseStatusAcceptsCurrentContract(t *testing.T) {
	if err := validateCurrentReleaseContract(currentReleaseStatusTestContract()); err != nil {
		t.Fatal(err)
	}
}

func TestCurrentReleaseStatusRequiresVerifiedRollingLineage(t *testing.T) {
	required := []string{
		"platform.schema_contracts AS contract",
		"platform.release_generations AS active_release",
		"platform.release_generations AS compatible_release",
		"platform.environment_identity AS identity",
		"platform.migration_runs AS run",
		"run.command='up' AND run.state='succeeded'",
		"active_release.bootstrap_receipt_run_id",
		"active_release.extension_receipt_run_id",
		"active_release.cleanup_receipt_run_id",
		"contract.database_generation >= $3",
		"$2 >= contract.minimum_runtime_generation",
		"compatible_release.runtime_generation=$2",
		"compatible_release.minimum_database_generation=$3",
		"compatible_release.migration_manifest_sha256=$4",
		"compatible_release.extension_manifest_sha256=$5",
		"compatible_release.postgres_image_digest=$6",
	}
	for _, fragment := range required {
		if !strings.Contains(currentReleaseAppliedSQL, fragment) {
			t.Fatalf("current release status omitted %q", fragment)
		}
	}
}

func currentReleaseStatusTestContract() runtimev1.BinarySchemaContract {
	return runtimev1.BinarySchemaContract{
		ContractVersion: runtimev1.BinarySchemaContractVersion,
		Component:       "core", RuntimeGeneration: 2, MinimumDatabaseGeneration: 2,
		MigrationManifestSHA256: strings.Repeat("a", 64),
		ExtensionManifestSHA256: strings.Repeat("b", 64),
		PostgresImageDigest:     runtimev1.PostgresImageDigest,
	}
}
