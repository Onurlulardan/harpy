package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

func runContract(arguments []string) error {
	return exportContract(arguments, loadMigrationArtifact)
}

func runFoundationContract(arguments []string) error {
	return exportContract(arguments, loadFoundationMigrationArtifact)
}

func exportContract(
	arguments []string,
	loadMigration func() (artifact, error),
) error {
	options, err := parseContractOptions(arguments)
	if err != nil {
		return err
	}
	if options.output == options.migrationOutput {
		return fmt.Errorf("contract and migration outputs must differ")
	}
	migration, err := loadMigration()
	if err != nil {
		return err
	}
	extension, err := exactExtensionArtifact()
	if err != nil {
		return err
	}
	contract, err := localSchemaContract(migration, extension.digest)
	if err != nil {
		return err
	}
	document, err := attestation.MarshalCanonical(contract)
	if err != nil {
		return fmt.Errorf("encode binary schema contract: %w", err)
	}
	if err := contract.Validate(); err != nil {
		return err
	}
	if err := writeExactCanonical(options.migrationOutput, migration.document); err != nil {
		return err
	}
	if err := writeExactCanonical(options.output, document); err != nil {
		return err
	}
	digest, _ := attestation.DigestHex(document)
	fmt.Printf("binary schema contract: %s\nsha256: %s\nmigration manifest: %s\nmigration_sha256: %s\n",
		filepath.Clean(options.output), digest, filepath.Clean(options.migrationOutput), migration.digest)
	return nil
}

func writeExactCanonical(path string, document []byte) error {
	existing, err := os.ReadFile(path)
	if err == nil {
		if bytes.Equal(existing, document) {
			return nil
		}
		return fmt.Errorf("generated output %s already exists with different content", path)
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("inspect generated output %s: %w", path, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create generated output parent: %w", err)
	}
	return attestation.WriteNewCanonicalBounded(
		path, document, maximumLocalReleaseArtifactBytes,
	)
}
