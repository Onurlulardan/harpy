package attestation

import (
	"fmt"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func BuildBinarySchemaContract(
	release VerifiedRelease,
	migration VerifiedPostgresArtifact[postgresv1.MigrationManifest],
	extension VerifiedPostgresArtifact[postgresv1.ExtensionManifest],
) (VerifiedBinarySchemaContract, error) {
	if release.ManifestSHA256() == "" || !migration.IsVerified() || !extension.IsVerified() {
		return VerifiedBinarySchemaContract{}, fmt.Errorf(
			"binary schema contract inputs are not verified",
		)
	}
	manifest := release.Value()
	if !sameDigest(manifest.MigrationManifestSHA256, migration.DigestSHA256()) ||
		!sameDigest(manifest.ExtensionManifestSHA256, extension.DigestSHA256()) {
		return VerifiedBinarySchemaContract{}, fmt.Errorf(
			"binary schema contract artifact digest differs from release",
		)
	}
	extensionManifest := extension.Value()
	if extensionManifest.DatabaseOrigin != manifest.DatabaseOrigin ||
		extensionManifest.PostgresImageDigest != manifest.PostgresImageDigest ||
		extensionManifest.PostgresEngineVersion != manifest.PostgresEngineVersion {
		return VerifiedBinarySchemaContract{}, fmt.Errorf(
			"binary schema contract extension identity differs from release",
		)
	}
	contract := runtimev1.BinarySchemaContract{
		ContractVersion: runtimev1.BinarySchemaContractVersion,
		Component:       "core", RuntimeGeneration: manifest.RuntimeGeneration,
		MinimumDatabaseGeneration: manifest.MinimumDatabaseGeneration,
		MigrationManifestSHA256:   migration.DigestSHA256(),
		ExtensionManifestSHA256:   extension.DigestSHA256(),
		PostgresImageDigest:       manifest.PostgresImageDigest,
	}
	document, err := MarshalCanonical(contract)
	if err != nil {
		return VerifiedBinarySchemaContract{}, fmt.Errorf(
			"encode binary schema contract: %w",
			err,
		)
	}
	return VerifyBinarySchemaContract(document)
}
