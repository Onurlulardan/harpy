package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

func TestFoundationContractUsesPinnedBootstrapBaseline(t *testing.T) {
	directory := t.TempDir()
	contractPath := filepath.Join(directory, "binary-schema-contract.json")
	migrationPath := filepath.Join(directory, "migration-manifest.json")
	if err := runFoundationContract([]string{
		"-output", contractPath, "-migration-output", migrationPath,
	}); err != nil {
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
	if contract.Value().MigrationManifestSHA256 != foundationMigrationDigest ||
		contract.Value().RuntimeGeneration != 1 ||
		contract.Value().MinimumDatabaseGeneration != 2 {
		t.Fatal("foundation binary contract differs from the pinned release")
	}
	migrationDocument, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatal(err)
	}
	migration, err := attestation.VerifyMigrationManifestArtifact(migrationDocument)
	if err != nil {
		t.Fatal(err)
	}
	if migration.DigestSHA256() != foundationMigrationDigest ||
		len(migration.Value().Entries) != 138 {
		t.Fatal("exported foundation migration artifact differs")
	}
}
