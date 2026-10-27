//go:build integration

package main

import (
	"strings"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func workflowBootstrapAuthority(now time.Time) runtimev1.BootstrapAuthorityManifest {
	template, digest := "template0", strings.Repeat("a", 64)
	roles := []runtimev1.BootstrapAuthorityRole{
		workflowBootstrapRole("harpy_owner", true, "permanent", false, false, false, false, -1),
		workflowBootstrapRole("harpy_cutover_authority_owner", true, "externally_provisioned", false, false, false, false, -1),
		workflowBootstrapRole("harpy_runtime", true, "permanent", false, false, false, false, -1),
		workflowBootstrapRole("harpy_v2_runtime", true, "permanent", true, false, false, false, -1),
		workflowBootstrapRole("harpy_migrator", true, "permanent", true, false, false, false, -1),
		workflowBootstrapRole("harpy_v1_runtime", false, "coexistence", false, false, false, false, -1),
		workflowBootstrapRole("harpy_cluster_admin", true, "leased", false, false, true, true, 1),
		workflowBootstrapRole("harpy_extension_admin", true, "leased", false, true, false, false, 1),
		workflowBootstrapRole("harpy_legacy_extension_owner", false, "adopted_only", false, false, false, false, -1),
		workflowBootstrapRole("harpy_adopter", true, "cutover_leased", false, false, false, false, 0),
		workflowBootstrapRole("harpy_cutover_decoder", true, "cutover_leased", false, false, false, false, 0),
		workflowBootstrapRole("harpy_cutover_finalizer", true, "cutover_finalization_leased", false, false, false, false, 0),
		workflowBootstrapRole("harpy_backup_reader", true, "permanent", false, false, false, false, -1),
		workflowBootstrapRole("harpy_backup", true, "permanent", true, false, false, false, -1),
		workflowBootstrapRole("harpy_restore", true, "permanent", true, false, false, false, -1),
		workflowBootstrapRole("harpy_reactivation_observer", true, "permanent", true, false, false, false, -1),
	}
	memberships := []runtimev1.BootstrapAuthorityMembership{
		workflowMembership("harpy_v2_runtime", "harpy_runtime", true, false, true),
		workflowMembership("harpy_migrator", "harpy_owner", true, false, true),
		workflowMembership("harpy_backup", "harpy_backup_reader", true, false, true),
		workflowMembership("harpy_restore", "harpy_owner", true, false, true),
		workflowMembership("harpy_v1_runtime", "harpy_runtime", false, true, false),
	}
	manifest := runtimev1.BootstrapAuthorityManifest{
		ContractVersion: "bootstrap-authority-manifest.v1",
		ManifestID:      "11111111-1111-4111-8111-111111111119",
		EnvironmentID:   workflowEnvironmentID, ReleaseID: workflowReleaseID,
		IssuedAt:       now.Add(-time.Minute).Format(time.RFC3339),
		ExpiresAt:      now.Add(10 * time.Minute).Format(time.RFC3339),
		DatabaseOrigin: "fresh", Database: runtimev1.BootstrapAuthorityDatabase{
			Name: workflowDatabaseName, CreateDatabase: true, Owner: "harpy_owner",
			ExtensionOwner: "harpy_extension_admin", ExtensionOwnerMode: "fresh_leased_admin",
			OwnershipExecutor: "cluster_bootstrap", OwnershipAttestation: "bootstrap_live_catalog",
			Encoding: runtimev1.PostgresEncoding, LocaleProvider: runtimev1.PostgresLocaleProvider,
			Collation: runtimev1.PostgresCollation, CType: runtimev1.PostgresCharacterType,
			DataChecksums: runtimev1.PostgresDataChecksums, Template: &template,
		},
		Connections: runtimev1.BootstrapAuthorityConnections{EndpointKind: "direct_postgres",
			MaintenanceDatabase: "postgres", TargetDatabase: workflowDatabaseName,
			SameServerRequired: true},
		Roles: roles, Memberships: memberships,
		CredentialLeases: []runtimev1.BootstrapCredentialLease{
			workflowBootstrapLease("11111111-1111-4111-8111-111111111117", "harpy_cluster_admin", now),
			workflowBootstrapLease("11111111-1111-4111-8111-111111111118", "harpy_extension_admin", now),
		},
		RolesManifestSHA256: digest, PrivilegesManifestSHA256: digest,
		OwnershipManifestSHA256: digest, BootstrapOperationDigest: digest,
		CredentialCleanupOperationDigest: digest,
	}
	return manifest
}

func workflowBootstrapRole(
	name string, enabled bool, mode string, login, superuser, createDB, createRole bool,
	limit int,
) runtimev1.BootstrapAuthorityRole {
	return runtimev1.BootstrapAuthorityRole{Name: name, Enabled: enabled,
		CredentialMode: mode, Login: login, Superuser: superuser, CreateDB: createDB,
		CreateRole: createRole, ConnectionLimit: limit}
}

func workflowMembership(
	member, role string, enabled, inherit, set bool,
) runtimev1.BootstrapAuthorityMembership {
	return runtimev1.BootstrapAuthorityMembership{Member: member, Role: role,
		Enabled: enabled, Inherit: inherit, Set: set, Grantor: "harpy_cluster_admin"}
}

func workflowBootstrapLease(
	id, principal string, now time.Time,
) runtimev1.BootstrapCredentialLease {
	return runtimev1.BootstrapCredentialLease{LeaseID: id, Principal: principal,
		NotBefore:          now.Add(-time.Minute).Format(time.RFC3339),
		ExpiresAt:          now.Add(10 * time.Minute).Format(time.RFC3339),
		MaxDurationSeconds: 900, MaxConnections: 1,
		CleanupPolicy: "disable_login_clear_password_terminate_sessions"}
}
