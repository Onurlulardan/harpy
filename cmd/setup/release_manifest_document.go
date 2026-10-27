package main

import (
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func buildReleaseManifest(
	keys releaseKeys, manifests postgresManifests, operations postgresOperations,
	authorityDigest, catalogDigest, cutoverDigest string,
) (releaseDocument, []byte, attestation.VerifiedRelease, error) {
	document, err := canonicalDocument(attestation.ReleaseManifest{
		ContractVersion: "postgres-release-manifest.v1", ReleaseID: localReleaseID,
		IssuedAt: keys.clock.stamp(keys.clock.issued), ExpiresAt: keys.clock.stamp(keys.clock.expires),
		DatabaseName: localDatabase, DatabaseOrigin: "fresh",
		DatabaseEncoding: runtimev1.PostgresEncoding, DatabaseLocaleProvider: runtimev1.PostgresLocaleProvider,
		DatabaseCollation: runtimev1.PostgresCollation, DatabaseCType: runtimev1.PostgresCharacterType,
		DatabaseOwner: runtimev1.DatabaseOwnerRole, DataChecksums: runtimev1.PostgresDataChecksums,
		PostgresImageDigest: runtimev1.PostgresImageDigest, PostgresEngineVersion: runtimev1.PostgresEngineVersion,
		VectorExtensionVersion: runtimev1.VectorExtensionVersion,
		RuntimeGeneration:      1, MinimumDatabaseGeneration: 1,
		MigrationManifestSHA256: manifests.migration.digest,
		ExtensionManifestSHA256: manifests.extension.digest,
		RolesManifestSHA256:     manifests.roles.digest, PrivilegesManifestSHA256: manifests.privileges.digest,
		OwnershipManifestSHA256: manifests.ownership.digest, BootstrapAuthoritySHA256: authorityDigest,
		BootstrapOperationDigest:        operations.bootstrap.digest,
		ExtensionOperationDigest:        operations.extension.digest,
		CredentialCleanupDigest:         operations.cleanup.digest,
		CutoverAuthorityOperationDigest: cutoverDigest, CatalogPackageSetSHA256: catalogDigest,
	})
	if err != nil {
		return releaseDocument{}, nil, attestation.VerifiedRelease{}, err
	}
	signature, err := keys.signDetached("release-signature.v1", document.body)
	if err != nil {
		return releaseDocument{}, nil, attestation.VerifiedRelease{}, err
	}
	verified, err := attestation.VerifyRelease(
		document.body, signature, keys.root, keys.clock.now, localDatabase,
	)
	if err != nil {
		return releaseDocument{}, nil, attestation.VerifiedRelease{}, err
	}
	return document, signature, verified, nil
}
