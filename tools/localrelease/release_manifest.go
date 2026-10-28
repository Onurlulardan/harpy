package main

import (
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func buildReleaseManifest(
	now time.Time, ids releaseIDs, catalog releaseCatalog,
	authority, cutover signedArtifactPair, signer signingAuthority,
) (signedArtifactPair, attestation.ReleaseManifest, error) {
	issuedAt := now.Add(-time.Minute)
	manifest := attestation.ReleaseManifest{
		ContractVersion: "postgres-release-manifest.v1", ReleaseID: ids.release,
		IssuedAt:     issuedAt.Format(time.RFC3339),
		ExpiresAt:    now.Add(24 * time.Hour).Format(time.RFC3339),
		DatabaseName: databaseName, DatabaseOrigin: "fresh",
		DatabaseEncoding:                runtimev1.PostgresEncoding,
		DatabaseLocaleProvider:          runtimev1.PostgresLocaleProvider,
		DatabaseCollation:               runtimev1.PostgresCollation,
		DatabaseCType:                   runtimev1.PostgresCharacterType,
		DatabaseOwner:                   runtimev1.DatabaseOwnerRole,
		DataChecksums:                   runtimev1.PostgresDataChecksums,
		PostgresImageDigest:             runtimev1.PostgresImageDigest,
		PostgresEngineVersion:           runtimev1.PostgresEngineVersion,
		VectorExtensionVersion:          runtimev1.VectorExtensionVersion,
		RuntimeGeneration:               catalog.schemaContract.RuntimeGeneration,
		MinimumDatabaseGeneration:       catalog.schemaContract.MinimumDatabaseGeneration,
		MigrationManifestSHA256:         catalog.migration.digest,
		ExtensionManifestSHA256:         catalog.extension.digest,
		RolesManifestSHA256:             catalog.roles.digest,
		PrivilegesManifestSHA256:        catalog.privileges.digest,
		OwnershipManifestSHA256:         catalog.ownership.digest,
		BootstrapAuthoritySHA256:        authority.document.digest,
		BootstrapOperationDigest:        catalog.bootstrapOperation.digest,
		ExtensionOperationDigest:        catalog.extensionOperation.digest,
		CredentialCleanupDigest:         catalog.cleanupOperation.digest,
		CutoverAuthorityOperationDigest: cutover.document.digest,
		CatalogPackageSetSHA256:         catalog.catalogPackageSet.digest,
	}
	document, err := newArtifact("release.json", manifest)
	if err != nil {
		return signedArtifactPair{}, manifest, err
	}
	signatureValue, err := runtimeSignature(signer, document.document, issuedAt)
	if err != nil {
		return signedArtifactPair{}, manifest, err
	}
	signed, err := newArtifact("release.sig", map[string]any{
		"contract_version": "release-signature.v1", "manifest_sha256": document.digest,
		"signature": signatureValue,
	})
	return signedArtifactPair{document: document, signature: signed}, manifest, err
}
