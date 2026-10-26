//go:build integration

package postgres

import (
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func writeBuiltCutoverRelease(
	t *testing.T, directory string, signer builtCutoverSigner, now time.Time,
	bootstrap, operation builtCutoverArtifact,
) (builtCutoverArtifact, builtCutoverArtifact) {
	t.Helper()
	issuedAt := now.Add(-time.Minute).UTC().Truncate(time.Second)
	digest := digestMarker('a')
	manifest := attestation.ReleaseManifest{
		ContractVersion: "postgres-release-manifest.v1", ReleaseID: integrationReleaseID,
		IssuedAt:     issuedAt.Format(time.RFC3339),
		ExpiresAt:    now.Add(9 * time.Minute).UTC().Truncate(time.Second).Format(time.RFC3339),
		DatabaseName: "harpy_cutover", DatabaseOrigin: "fresh",
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
		MigrationManifestSHA256: digestMarker('6'), ExtensionManifestSHA256: digest,
		RolesManifestSHA256: digest, PrivilegesManifestSHA256: digest,
		OwnershipManifestSHA256: digest, BootstrapAuthoritySHA256: bootstrap.digest,
		BootstrapOperationDigest: digest, ExtensionOperationDigest: digest,
		CredentialCleanupDigest:         digest,
		CutoverAuthorityOperationDigest: operation.digest,
		CatalogPackageSetSHA256:         digest,
	}
	document := writeBuiltCutoverArtifact(t, directory, "release.json", manifest)
	signed := writeBuiltCutoverArtifact(t, directory, "release.sig", map[string]any{
		"contract_version": "release-signature.v1", "manifest_sha256": document.digest,
		"signature": signBuiltCutoverDocument(t, signer, document.document, issuedAt),
	})
	return document, signed
}
