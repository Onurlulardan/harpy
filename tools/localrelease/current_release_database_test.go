package main

// PRESERVE_NON_MIGRATION: asserts release status does not require a Goose version row.

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

func TestCurrentReleaseStatusUsesBaselineGooseVersion(t *testing.T) {
	for _, removed := range []string{
		"platform.release_generations",
		"platform.schema_contracts",
		"platform.migration_runs",
		"platform.bootstrap_receipts",
	} {
		if strings.Contains(currentReleaseAppliedSQL, removed) {
			t.Fatalf("current release status still reads %s", removed)
		}
	}
	if !strings.Contains(currentReleaseAppliedSQL, "platform.environment_identity AS identity") {
		t.Fatal("current release status omitted the environment identity")
	}
	if strings.Contains(currentReleaseAppliedSQL, "goose_db_version") ||
		strings.Contains(currentReleaseAppliedSQL, "20261103000006") {
		t.Fatal("current release status still requires a Goose version")
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
