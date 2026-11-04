package main

import (
	"fmt"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func buildAuthority(
	keys releaseKeys, manifests postgresManifests, operations postgresOperations,
) (releaseDocument, []byte, error) {
	template := "template0"
	manifest, err := canonicalDocument(runtimev1.BootstrapAuthorityManifest{
		ContractVersion: "bootstrap-authority-manifest.v1", ManifestID: localAuthorityID,
		EnvironmentID: localEnvironmentID, ReleaseID: localReleaseID,
		IssuedAt: keys.clock.stamp(keys.clock.issued), ExpiresAt: keys.clock.stamp(keys.clock.releaseExpires),
		DatabaseOrigin: "fresh",
		Database: runtimev1.BootstrapAuthorityDatabase{
			Name: localDatabase, CreateDatabase: true, Owner: "harpy_owner",
			ExtensionOwner: "harpy_extension_admin", ExtensionOwnerMode: "fresh_leased_admin",
			OwnershipExecutor: "cluster_bootstrap", OwnershipAttestation: "bootstrap_live_catalog",
			Encoding: runtimev1.PostgresEncoding, LocaleProvider: runtimev1.PostgresLocaleProvider,
			Collation: runtimev1.PostgresCollation, CType: runtimev1.PostgresCharacterType,
			DataChecksums: runtimev1.PostgresDataChecksums, Template: &template,
		},
		Connections: runtimev1.BootstrapAuthorityConnections{
			EndpointKind: "direct_postgres", MaintenanceDatabase: "postgres",
			TargetDatabase: localDatabase, SameServerRequired: true,
		},
		Roles:       runtimev1.BootstrapAuthorityRoles(),
		Memberships: runtimev1.BootstrapAuthorityMemberships(),
		CredentialLeases: []runtimev1.BootstrapCredentialLease{
			authorityLease(keys, localClusterLease, "harpy_cluster_admin"),
			authorityLease(keys, localExtensionLease, "harpy_extension_admin"),
		},
		RolesManifestSHA256: manifests.roles.digest, PrivilegesManifestSHA256: manifests.privileges.digest,
		OwnershipManifestSHA256: manifests.ownership.digest, BootstrapOperationDigest: operations.bootstrap.digest,
		CredentialCleanupOperationDigest: operations.cleanup.digest,
	})
	if err != nil {
		return releaseDocument{}, nil, err
	}
	if err = runtimev1.ValidateBootstrapAuthorityManifest(authorityValue(manifest.body)); err != nil {
		return releaseDocument{}, nil, fmt.Errorf("bootstrap authority: %w", err)
	}
	signature, err := keys.signDetached("bootstrap-authority-signature.v1", manifest.body)
	if err != nil {
		return releaseDocument{}, nil, err
	}
	if _, err = attestation.VerifyBootstrapAuthority(
		manifest.body, signature, keys.root, keys.clock.now,
	); err != nil {
		return releaseDocument{}, nil, err
	}
	return manifest, signature, nil
}

func authorityValue(document []byte) runtimev1.BootstrapAuthorityManifest {
	var manifest runtimev1.BootstrapAuthorityManifest
	if err := attestation.DecodeStrict(document, &manifest); err != nil {
		return runtimev1.BootstrapAuthorityManifest{}
	}
	return manifest
}

func authorityLease(keys releaseKeys, id, principal string) runtimev1.BootstrapCredentialLease {
	return runtimev1.BootstrapCredentialLease{
		LeaseID: id, Principal: principal,
		NotBefore: keys.clock.stamp(keys.clock.issued), ExpiresAt: keys.clock.stamp(keys.clock.expires),
		MaxDurationSeconds: 900, MaxConnections: 1,
		CleanupPolicy: "disable_login_clear_password_terminate_sessions",
	}
}
