package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

func TestPrepareRequiresEveryAbsolutePath(t *testing.T) {
	root := t.TempDir()
	arguments := []string{
		"-admin-dsn-file", filepath.Join(root, "admin"),
		"-runtime-password-file", filepath.Join(root, "runtime"),
		"-migrator-password-file", filepath.Join(root, "migrator"),
		"-backup-password-file", filepath.Join(root, "backup"),
		"-restore-password-file", filepath.Join(root, "restore"),
		"-reactivation-observer-password-file", filepath.Join(root, "observer"),
		"-migration-manifest", filepath.Join(root, "migration.json"),
		"-catalog-package-signing-seed-file", filepath.Join(root, "catalog.seed"),
		"-catalog-package-keyring-file", filepath.Join(root, "catalog-keyring.json"),
		"-output", filepath.Join(root, "release"),
	}
	options, err := parsePrepareOptions(arguments)
	if err != nil {
		t.Fatal(err)
	}
	if options.outputDirectory != filepath.Join(root, "release") {
		t.Fatalf("output = %s", options.outputDirectory)
	}
	arguments[1] = "relative"
	if _, err := parsePrepareOptions(arguments); err == nil {
		t.Fatal("relative admin DSN path accepted")
	}
}

func TestContractExportsCanonicalBuildInputsIdempotently(t *testing.T) {
	directory := t.TempDir()
	contractPath := filepath.Join(directory, "binary-schema-contract.json")
	migrationPath := filepath.Join(directory, "migration-manifest.json")
	arguments := []string{
		"-output", contractPath, "-migration-output", migrationPath,
	}
	if err := runContract(arguments); err != nil {
		t.Fatal(err)
	}
	if err := runContract(arguments); err != nil {
		t.Fatal(err)
	}
	contractDocument, err := os.ReadFile(contractPath)
	if err != nil {
		t.Fatal(err)
	}
	contract, err := attestation.VerifyBinarySchemaContract(contractDocument)
	if err != nil {
		t.Fatal(err)
	}
	migrationDocument, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatal(err)
	}
	migration, err := attestation.VerifyMigrationManifestArtifact(migrationDocument)
	if err != nil {
		t.Fatal(err)
	}
	if contract.Value().MigrationManifestSHA256 != migration.DigestSHA256() {
		t.Fatal("exported contract and migration manifest differ")
	}
}

func TestRuntimeReadinessAllowlistIsExact(t *testing.T) {
	expected := []string{
		"environment_identity", "bootstrap_receipts", "extension_registry",
		"migration_runs", "schema_contracts", "database_generation_manifests",
		"release_generations",
	}
	if len(runtimeReadinessTables) != len(expected) {
		t.Fatalf("readiness table count = %d", len(runtimeReadinessTables))
	}
	for index := range expected {
		if runtimeReadinessTables[index] != expected[index] {
			t.Fatalf("readiness table %d = %s", index, runtimeReadinessTables[index])
		}
	}
}
