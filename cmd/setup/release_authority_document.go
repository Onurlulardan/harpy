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
		IssuedAt: keys.clock.stamp(keys.clock.issued), ExpiresAt: keys.clock.stamp(keys.clock.expires),
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
		Roles: authorityRoles(), Memberships: authorityMemberships(),
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

func authorityMemberships() []runtimev1.BootstrapAuthorityMembership {
	return []runtimev1.BootstrapAuthorityMembership{
		authorityMembership("harpy_v2_runtime", "harpy_runtime", true, false, true),
		authorityMembership("harpy_migrator", "harpy_owner", true, false, true),
		authorityMembership("harpy_backup", "harpy_backup_reader", true, false, true),
		authorityMembership("harpy_restore", "harpy_owner", true, false, true),
		authorityMembership("harpy_v1_runtime", "harpy_runtime", false, true, false),
	}
}

func authorityMembership(
	member, role string, enabled, inherit, set bool,
) runtimev1.BootstrapAuthorityMembership {
	return runtimev1.BootstrapAuthorityMembership{
		Member: member, Role: role, Enabled: enabled, Inherit: inherit, Set: set,
		Grantor: "harpy_cluster_admin",
	}
}

func authorityRoles() []runtimev1.BootstrapAuthorityRole {
	type spec struct {
		name, mode                                      string
		enabled, login, superuser, createDB, createRole bool
		limit                                           int
	}
	specs := []spec{
		{"harpy_owner", "permanent", true, false, false, false, false, -1},
		{"harpy_cutover_authority_owner", "externally_provisioned", true, false, false, false, false, -1},
		{"harpy_runtime", "permanent", true, false, false, false, false, -1},
		{"harpy_v2_runtime", "permanent", true, true, false, false, false, -1},
		{"harpy_migrator", "permanent", true, true, false, false, false, -1},
		{"harpy_v1_runtime", "coexistence", false, false, false, false, false, -1},
		{"harpy_cluster_admin", "leased", true, false, false, true, true, 1},
		{"harpy_extension_admin", "leased", true, false, true, false, false, 1},
		{"harpy_legacy_extension_owner", "adopted_only", false, false, false, false, false, -1},
		{"harpy_adopter", "cutover_leased", true, false, false, false, false, 0},
		{"harpy_cutover_decoder", "cutover_leased", true, false, false, false, false, 0},
		{"harpy_cutover_finalizer", "cutover_finalization_leased", true, false, false, false, false, 0},
		{"harpy_backup_reader", "permanent", true, false, false, false, false, -1},
		{"harpy_backup", "permanent", true, true, false, false, false, -1},
		{"harpy_restore", "permanent", true, true, false, false, false, -1},
		{"harpy_reactivation_observer", "permanent", true, true, false, false, false, -1},
	}
	roles := make([]runtimev1.BootstrapAuthorityRole, 0, len(specs))
	for _, item := range specs {
		roles = append(roles, runtimev1.BootstrapAuthorityRole{
			Name: item.name, Enabled: item.enabled, CredentialMode: item.mode, Login: item.login,
			Superuser: item.superuser, CreateDB: item.createDB, CreateRole: item.createRole,
			ConnectionLimit: item.limit,
		})
	}
	return roles
}
