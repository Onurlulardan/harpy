//go:build integration

package schemacontract

import (
	"fmt"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/binarycontract"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func IntegrationMigrationManifestSHA256() (string, error) {
	registry, err := migration.LoadReleaseRegistry()
	if err != nil {
		return "", fmt.Errorf("load integration migration registry: %w", err)
	}
	digests := registry.Digests()
	if len(digests) == 0 {
		return "", fmt.Errorf("integration migration registry is empty")
	}
	return digests[0], nil
}

func IntegrationContract(extensionManifestSHA256 string) (EmbeddedContract, error) {
	migrationManifestSHA256, err := IntegrationMigrationManifestSHA256()
	if err != nil {
		return EmbeddedContract{}, err
	}
	value := runtimev1.BinarySchemaContract{
		ContractVersion:           runtimev1.BinarySchemaContractVersion,
		Component:                 "core",
		RuntimeGeneration:         1,
		MinimumDatabaseGeneration: 1,
		MigrationManifestSHA256:   migrationManifestSHA256,
		ExtensionManifestSHA256:   extensionManifestSHA256,
		PostgresImageDigest:       runtimev1.PostgresImageDigest,
	}
	document, err := attestation.MarshalCanonical(value)
	if err != nil {
		return EmbeddedContract{}, fmt.Errorf("marshal integration schema contract: %w", err)
	}
	artifact, err := binarycontract.Verify(document)
	if err != nil {
		return EmbeddedContract{}, fmt.Errorf("verify integration schema contract: %w", err)
	}
	registry, err := migration.LoadReleaseRegistry()
	if err != nil {
		return EmbeddedContract{}, fmt.Errorf("reload integration migration registry: %w", err)
	}
	contract, err := authorizeEmbedded(registry, artifact)
	if err != nil {
		return EmbeddedContract{}, fmt.Errorf("authorize integration schema contract: %w", err)
	}
	return contract, nil
}
