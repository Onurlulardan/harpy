//go:build integration

package migrator

import (
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func writeBuiltEntrypointRelease(
	t *testing.T,
	directory string,
	now time.Time,
	catalog builtEntrypointArtifactCatalog,
	authority builtEntrypointArtifact,
	signer builtEntrypointSigner,
) (builtEntrypointArtifact, builtEntrypointArtifact) {
	t.Helper()
	manifest := attestation.ReleaseManifest{
		ContractVersion: "postgres-release-manifest.v1",
		ReleaseID:       builtEntrypointReleaseID,
		IssuedAt:        now.Add(-time.Minute).Format(time.RFC3339),
		ExpiresAt:       now.Add(9 * time.Minute).Format(time.RFC3339),
		DatabaseName:    builtEntrypointDatabaseName, DatabaseOrigin: "fresh",
		DatabaseEncoding:       runtimev1.PostgresEncoding,
		DatabaseLocaleProvider: runtimev1.PostgresLocaleProvider,
		DatabaseCollation:      runtimev1.PostgresCollation,
		DatabaseCType:          runtimev1.PostgresCharacterType,
		DatabaseOwner:          runtimev1.DatabaseOwnerRole,
		DataChecksums:          runtimev1.PostgresDataChecksums,
		PostgresImageDigest:    runtimev1.PostgresImageDigest,
		PostgresEngineVersion:  runtimev1.PostgresEngineVersion,
		VectorExtensionVersion: runtimev1.VectorExtensionVersion,
		RuntimeGeneration:      1, MinimumDatabaseGeneration: 1,
		MigrationManifestSHA256:         catalog.migration.digest,
		ExtensionManifestSHA256:         catalog.extension.digest,
		RolesManifestSHA256:             catalog.roles.digest,
		PrivilegesManifestSHA256:        catalog.privileges.digest,
		OwnershipManifestSHA256:         catalog.ownership.digest,
		BootstrapAuthoritySHA256:        authority.digest,
		BootstrapOperationDigest:        catalog.bootstrapOperation.digest,
		ExtensionOperationDigest:        catalog.extensionOperation.digest,
		CredentialCleanupDigest:         catalog.cleanupOperation.digest,
		CutoverAuthorityOperationDigest: catalog.cutoverAuthorityOperation.digest,
		CatalogPackageSetSHA256:         catalog.catalogPackageSet.digest,
	}
	document := writeBuiltEntrypointArtifact(t, directory, "release.json", manifest)
	signed := writeBuiltEntrypointArtifact(t, directory, "release.sig", map[string]any{
		"contract_version": "release-signature.v1",
		"manifest_sha256":  document.digest,
		"signature": builtEntrypointRuntimeSignature(
			t, signer.signer, document.document, now.Add(-time.Minute),
		),
	})
	return document, signed
}
