package main

import (
	"fmt"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

// A local release is a fresh, single-version stack rather than a rolling
// production release. Its binary therefore requires the canonical migration
// frontier and advertises the runtime generation required by that frontier.
func localSchemaContract(
	migration artifact,
	extensionDigest string,
) (runtimev1.BinarySchemaContract, error) {
	verified, err := attestation.VerifyMigrationManifestArtifact(migration.document)
	if err != nil {
		return runtimev1.BinarySchemaContract{}, fmt.Errorf(
			"verify local migration frontier: %w", err,
		)
	}
	if migration.digest != verified.DigestSHA256() {
		return runtimev1.BinarySchemaContract{}, fmt.Errorf(
			"local migration frontier digest differs",
		)
	}
	entries := verified.Value().Entries
	if len(entries) == 0 {
		return runtimev1.BinarySchemaContract{}, fmt.Errorf(
			"local migration frontier is empty",
		)
	}
	frontier := entries[len(entries)-1]
	contract := runtimev1.BinarySchemaContract{
		ContractVersion:           runtimev1.BinarySchemaContractVersion,
		Component:                 "core",
		RuntimeGeneration:         frontier.NextMinimumRuntimeGeneration,
		MinimumDatabaseGeneration: frontier.TargetDatabaseGeneration,
		MigrationManifestSHA256:   verified.DigestSHA256(),
		ExtensionManifestSHA256:   extensionDigest,
		PostgresImageDigest:       runtimev1.PostgresImageDigest,
	}
	if err := contract.Validate(); err != nil {
		return runtimev1.BinarySchemaContract{}, fmt.Errorf(
			"validate local schema contract: %w", err,
		)
	}
	return contract, nil
}

func freshBinarySchemaContract(
	migration artifact, _ string,
) (runtimev1.BinarySchemaContract, error) {
	contract, err := readEmbeddedBuildContract()
	if err != nil {
		return runtimev1.BinarySchemaContract{}, err
	}
	contract.MigrationManifestSHA256 = migration.digest
	if err := contract.Validate(); err != nil {
		return runtimev1.BinarySchemaContract{}, fmt.Errorf(
			"validate local schema contract: %w", err,
		)
	}
	return contract, nil
}
