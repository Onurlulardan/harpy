package main

import (
	"fmt"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

type signedArtifactPair struct {
	document, signature artifact
}

func buildBootstrapAuthority(
	now time.Time, ids releaseIDs, catalog releaseCatalog, signer signingAuthority,
) (signedArtifactPair, error) {
	issuedAt, expiresAt := now.Add(-time.Minute), now.Add(14*time.Minute)
	template := "template0"
	manifest := runtimev1.BootstrapAuthorityManifest{
		ContractVersion: "bootstrap-authority-manifest.v1",
		ManifestID:      ids.authorityManifest, EnvironmentID: ids.environment,
		ReleaseID: ids.release, IssuedAt: issuedAt.Format(time.RFC3339),
		ExpiresAt: expiresAt.Format(time.RFC3339), DatabaseOrigin: "fresh",
		Database: runtimev1.BootstrapAuthorityDatabase{
			Name: databaseName, CreateDatabase: true, Owner: "harpy_owner",
			ExtensionOwner:       "harpy_extension_admin",
			ExtensionOwnerMode:   "fresh_leased_admin",
			OwnershipExecutor:    "cluster_bootstrap",
			OwnershipAttestation: "bootstrap_live_catalog",
			Encoding:             runtimev1.PostgresEncoding,
			LocaleProvider:       runtimev1.PostgresLocaleProvider,
			Collation:            runtimev1.PostgresCollation, CType: runtimev1.PostgresCharacterType,
			DataChecksums: runtimev1.PostgresDataChecksums, Template: &template,
		},
		Connections: runtimev1.BootstrapAuthorityConnections{
			EndpointKind: "direct_postgres", MaintenanceDatabase: "postgres",
			TargetDatabase: databaseName, SameServerRequired: true,
		},
		Roles: authorityRoles(), Memberships: authorityMemberships(),
		CredentialLeases: []runtimev1.BootstrapCredentialLease{
			authorityLease(ids.clusterLease, "harpy_cluster_admin", issuedAt, expiresAt),
			authorityLease(ids.extensionLease, "harpy_extension_admin", issuedAt, expiresAt),
		},
		RolesManifestSHA256:              catalog.roles.digest,
		PrivilegesManifestSHA256:         catalog.privileges.digest,
		OwnershipManifestSHA256:          catalog.ownership.digest,
		BootstrapOperationDigest:         catalog.bootstrapOperation.digest,
		CredentialCleanupOperationDigest: catalog.cleanupOperation.digest,
	}
	if err := runtimev1.ValidateBootstrapAuthorityManifest(manifest); err != nil {
		return signedArtifactPair{}, fmt.Errorf("validate bootstrap authority: %w", err)
	}
	document, err := newArtifact("bootstrap-authority.json", manifest)
	if err != nil {
		return signedArtifactPair{}, err
	}
	signatureValue, err := runtimeSignature(signer, document.document, issuedAt)
	if err != nil {
		return signedArtifactPair{}, err
	}
	signature := runtimev1.BootstrapAuthoritySignature{
		ContractVersion: "bootstrap-authority-signature.v1",
		ManifestSHA256:  document.digest, Signature: signatureValue,
	}
	if err := runtimev1.ValidateBootstrapAuthoritySignature(signature); err != nil {
		return signedArtifactPair{}, err
	}
	signed, err := newArtifact("bootstrap-authority.sig", signature)
	return signedArtifactPair{document: document, signature: signed}, err
}

func authorityRoles() []runtimev1.BootstrapAuthorityRole {
	return []runtimev1.BootstrapAuthorityRole{
		authorityRole("harpy_owner", true, "permanent", false, false, false, false, -1),
		authorityRole("harpy_cutover_authority_owner", true, "externally_provisioned", false, false, false, false, -1),
		authorityRole("harpy_runtime", true, "permanent", false, false, false, false, -1),
		authorityRole("harpy_v2_runtime", true, "permanent", true, false, false, false, -1),
		authorityRole("harpy_migrator", true, "permanent", true, false, false, false, -1),
		authorityRole("harpy_v1_runtime", false, "coexistence", false, false, false, false, -1),
		authorityRole("harpy_cluster_admin", true, "leased", false, false, true, true, 1),
		authorityRole("harpy_extension_admin", true, "leased", false, true, false, false, 1),
		authorityRole("harpy_legacy_extension_owner", false, "adopted_only", false, false, false, false, -1),
		authorityRole("harpy_adopter", true, "cutover_leased", false, false, false, false, 0),
		authorityRole("harpy_cutover_decoder", true, "cutover_leased", false, false, false, false, 0),
		authorityRole("harpy_cutover_finalizer", true, "cutover_finalization_leased", false, false, false, false, 0),
		authorityRole("harpy_backup_reader", true, "permanent", false, false, false, false, -1),
		authorityRole("harpy_backup", true, "permanent", true, false, false, false, -1),
		authorityRole("harpy_restore", true, "permanent", true, false, false, false, -1),
		authorityRole("harpy_reactivation_observer", true, "permanent", true, false, false, false, -1),
	}
}

func authorityRole(
	name string, enabled bool, mode string, login, superuser, createDB, createRole bool, limit int,
) runtimev1.BootstrapAuthorityRole {
	return runtimev1.BootstrapAuthorityRole{
		Name: name, Enabled: enabled, CredentialMode: mode, Login: login,
		Superuser: superuser, CreateDB: createDB, CreateRole: createRole,
		ConnectionLimit: limit,
	}
}

func authorityMemberships() []runtimev1.BootstrapAuthorityMembership {
	return []runtimev1.BootstrapAuthorityMembership{
		authorityMembership("harpy_v2_runtime", "harpy_runtime", true, false, true),
		authorityMembership("harpy_migrator", "harpy_owner", true, false, true),
		authorityMembership("harpy_backup", "harpy_backup_reader", true, false, true),
		authorityMembership("harpy_restore", "harpy_owner", true, false, true),
		authorityMembership("harpy_v1_runtime", "harpy_runtime", false, true, false),
	}
}

func authorityMembership(member, role string, enabled, inherit, set bool) runtimev1.BootstrapAuthorityMembership {
	return runtimev1.BootstrapAuthorityMembership{
		Member: member, Role: role, Enabled: enabled, Inherit: inherit, Set: set,
		Grantor: "harpy_cluster_admin",
	}
}

func authorityLease(
	id, principal string, issuedAt, expiresAt time.Time,
) runtimev1.BootstrapCredentialLease {
	return runtimev1.BootstrapCredentialLease{
		LeaseID: id, Principal: principal, NotBefore: issuedAt.Format(time.RFC3339),
		ExpiresAt: expiresAt.Format(time.RFC3339), MaxDurationSeconds: 900,
		MaxConnections: 1, CleanupPolicy: "disable_login_clear_password_terminate_sessions",
	}
}
