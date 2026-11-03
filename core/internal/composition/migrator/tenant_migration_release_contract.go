package migrator

import (
	"fmt"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func validateTenantMigrationReleaseContract(
	manifest attestation.ReleaseManifest,
	binary runtimev1.BinarySchemaContract,
) error {
	if manifest.ReleaseID == "" ||
		manifest.RuntimeGeneration != binary.RuntimeGeneration ||
		manifest.MinimumDatabaseGeneration != binary.MinimumDatabaseGeneration ||
		manifest.MigrationManifestSHA256 != binary.MigrationManifestSHA256 ||
		manifest.ExtensionManifestSHA256 != binary.ExtensionManifestSHA256 ||
		manifest.PostgresImageDigest != binary.PostgresImageDigest ||
		manifest.PostgresEngineVersion != runtimev1.PostgresEngineVersion ||
		manifest.VectorExtensionVersion != runtimev1.VectorExtensionVersion {
		return fmt.Errorf("signed tenant migration release does not match the embedded binary schema contract")
	}
	return nil
}
