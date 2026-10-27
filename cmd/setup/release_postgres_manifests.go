package main

import (
	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type postgresManifests struct {
	roles, privileges, ownership, extension, migration releaseDocument
}

func buildPostgresManifests() (postgresManifests, error) {
	roles, err := canonicalDocument(localRolesManifest())
	if err != nil {
		return postgresManifests{}, err
	}
	if _, err = attestation.VerifyRolesManifestArtifact(roles.body); err != nil {
		return postgresManifests{}, err
	}
	privileges, err := canonicalDocument(localPrivilegesManifest())
	if err != nil {
		return postgresManifests{}, err
	}
	if _, err = attestation.VerifyPrivilegesManifestArtifact(privileges.body); err != nil {
		return postgresManifests{}, err
	}
	ownership, err := canonicalDocument(localOwnershipManifest())
	if err != nil {
		return postgresManifests{}, err
	}
	if _, err = attestation.VerifyOwnershipManifestArtifact(ownership.body); err != nil {
		return postgresManifests{}, err
	}
	extension, err := canonicalDocument(localExtensionManifest())
	if err != nil {
		return postgresManifests{}, err
	}
	if _, err = attestation.VerifyExtensionManifestArtifact(extension.body); err != nil {
		return postgresManifests{}, err
	}
	migration, err := canonicalDocument(localMigrationManifest())
	if err != nil {
		return postgresManifests{}, err
	}
	if _, err = attestation.VerifyMigrationManifestArtifact(migration.body); err != nil {
		return postgresManifests{}, err
	}
	return postgresManifests{
		roles: roles, privileges: privileges, ownership: ownership,
		extension: extension, migration: migration,
	}, nil
}

func localRolesManifest() postgresv1.RolesManifest {
	names := []string{
		"harpy_owner", "harpy_cutover_authority_owner", "harpy_runtime", "harpy_v2_runtime",
		"harpy_migrator", "harpy_cluster_admin", "harpy_extension_admin", "harpy_adopter",
		"harpy_cutover_decoder", "harpy_cutover_finalizer", "harpy_backup_reader",
		"harpy_backup", "harpy_restore", "harpy_reactivation_observer",
		"harpy_tenant_migration_executor",
	}
	roles := make([]postgresv1.PostgresRole, 0, len(names))
	for _, name := range names {
		role := postgresv1.PostgresRole{
			Name: name, ConnectionLimit: -1, PasswordState: "null", ValidUntil: "infinity",
		}
		switch name {
		case "harpy_v2_runtime", "harpy_migrator", "harpy_backup",
			"harpy_restore", "harpy_reactivation_observer":
			role.Login = true
		case "harpy_cluster_admin":
			role.CreateDB, role.CreateRole, role.ConnectionLimit = true, true, 1
		case "harpy_extension_admin":
			role.Superuser, role.ConnectionLimit = true, 1
		case "harpy_adopter", "harpy_cutover_decoder", "harpy_cutover_finalizer":
			role.ConnectionLimit = 0
		}
		roles = append(roles, role)
	}
	return postgresv1.RolesManifest{
		ContractVersion: "postgres-roles-manifest.v1", DatabaseOrigin: "fresh", Roles: roles,
	}
}

func localPrivilegesManifest() postgresv1.PrivilegesManifest {
	grantees := []string{
		"harpy_backup", "harpy_cluster_admin", "harpy_extension_admin", "harpy_migrator",
		"harpy_reactivation_observer", "harpy_restore", "harpy_runtime", "harpy_v2_runtime",
	}
	grants := make([]postgresv1.ACLGrant, 0, len(grantees))
	for _, grantee := range grantees {
		grants = append(grants, postgresv1.ACLGrant{
			Object: postgresv1.CatalogObject{
				Kind: "database", DatabaseName: localDatabase, ObjectName: localDatabase,
			},
			Grantor: "harpy_owner", Grantee: grantee, Privileges: []string{"CONNECT"},
		})
	}
	return postgresv1.PrivilegesManifest{
		ContractVersion: "postgres-privileges-manifest.v1", DatabaseOrigin: "fresh",
		DatabaseName: localDatabase, PublicDatabasePrivileges: []string{},
		Memberships: []postgresv1.RoleMembership{
			{Member: "harpy_v2_runtime", Role: "harpy_runtime", SetOption: true},
			{Member: "harpy_migrator", Role: "harpy_owner", SetOption: true},
			{Member: "harpy_backup", Role: "harpy_backup_reader", SetOption: true},
			{Member: "harpy_restore", Role: "harpy_owner", SetOption: true},
		},
		Grants: grants, DefaultPrivileges: []postgresv1.DefaultPrivilege{},
		ExecutorDecisions: postgresv1.ExecutorDecisionSet{
			BootstrapRole: "harpy_cluster_admin", ExtensionRole: "harpy_extension_admin",
			RuntimeSessionRole: "harpy_v2_runtime", RuntimeCurrentRole: "harpy_runtime",
			MigratorSessionRole: "harpy_migrator", MigratorCurrentRole: "harpy_owner",
			OwnershipExecutor: "cluster_bootstrap",
		},
	}
}

func localOwnershipManifest() postgresv1.OwnershipManifest {
	return postgresv1.OwnershipManifest{
		ContractVersion: "postgres-ownership-manifest.v1", ManifestID: localOwnershipID,
		DatabaseOrigin: "fresh", DatabaseName: localDatabase,
		OwnershipExecutor: "cluster_bootstrap", Format: "inline", ObjectCount: 1,
		Objects: []postgresv1.OwnershipObject{{
			Object: postgresv1.CatalogObject{
				Kind: "database", DatabaseName: localDatabase, ObjectName: localDatabase,
			},
			Owner: "harpy_owner", Classification: "first_party", Executor: "cluster_bootstrap",
		}},
	}
}

func localExtensionManifest() postgresv1.ExtensionManifest {
	return postgresv1.ExtensionManifest{
		ContractVersion: "postgres-extension-manifest.v1", DatabaseOrigin: "fresh",
		PostgresImageDigest:   postgresv1.PostgresImageDigest,
		PostgresEngineVersion: postgresv1.PostgresEngineVersion,
		ControlPathPolicy: postgresv1.ControlPathPolicy{
			Resolver: "pg_config_sharedir", RelativeDirectory: "extension",
		},
		Extensions: []postgresv1.PostgresExtension{
			localExtension("pg_trgm", "1.6", "trigram_search"),
			localExtension("vector", "0.8.3", "vector_similarity"),
		},
	}
}

func localExtension(name, version, capability string) postgresv1.PostgresExtension {
	return postgresv1.PostgresExtension{
		Name: name, Version: version, Install: true, Schema: "platform_ext",
		Owner: "harpy_extension_admin", ControlFile: name + ".control",
		ControlFileSHA256: localFixedDigest, Capabilities: []string{capability},
	}
}

func localMigrationManifest() postgresv1.MigrationManifest {
	return postgresv1.MigrationManifest{
		ContractVersion: "postgres-migration-manifest.v1", Engine: "goose.v3",
		PostgresMajor: 18, HistoryTable: "platform.platform_migration_history",
		Entries: []postgresv1.MigrationEntry{{
			Owner: "platform", Scope: "platform", Version: 20260713010100,
			TransactionMode: "transactional", Phase: "foundation",
			TargetDatabaseGeneration: 1, NextMinimumRuntimeGeneration: 1,
			Variants: postgresv1.MigrationVariants{Fresh: &postgresv1.MigrationVariant{
				Path:   "core/internal/platform/migrator/migrations/fresh/20260713010100_local_release.fresh.sql",
				SHA256: localFixedDigest,
			}},
		}},
	}
}
