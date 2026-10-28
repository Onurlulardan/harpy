package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

func TestLocalSchemaContractTracksCanonicalMigrationFrontier(t *testing.T) {
	migration, err := loadMigrationArtifact()
	if err != nil {
		t.Fatal(err)
	}
	contract, err := localSchemaContract(migration, extensionManifestHash)
	if err != nil {
		t.Fatal(err)
	}
	verified, err := attestation.VerifyMigrationManifestArtifact(migration.document)
	if err != nil {
		t.Fatal(err)
	}
	entries := verified.Value().Entries
	frontier := entries[len(entries)-1]
	if contract.RuntimeGeneration != frontier.NextMinimumRuntimeGeneration ||
		contract.MinimumDatabaseGeneration != frontier.TargetDatabaseGeneration {
		t.Fatalf(
			"contract generations=(%d,%d), frontier=(%d,%d)",
			contract.RuntimeGeneration, contract.MinimumDatabaseGeneration,
			frontier.NextMinimumRuntimeGeneration, frontier.TargetDatabaseGeneration,
		)
	}
	embedded, err := readEmbeddedBuildContract()
	if err != nil {
		t.Fatal(err)
	}
	if contract != embedded {
		t.Fatal("fresh frontier contract differs from the embedded binary schema contract")
	}
}

func TestPreparedMigrationRejectsGenerationMismatch(t *testing.T) {
	migration, err := loadMigrationArtifact()
	if err != nil {
		t.Fatal(err)
	}
	contract, err := localSchemaContract(migration, extensionManifestHash)
	if err != nil {
		t.Fatal(err)
	}
	contract.RuntimeGeneration--
	path := filepath.Join(t.TempDir(), "migration-manifest.json")
	if err := os.WriteFile(path, migration.document, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadPreparedMigration(path, contract); err == nil {
		t.Fatal("generation-mismatched embedded contract was accepted")
	}
}
