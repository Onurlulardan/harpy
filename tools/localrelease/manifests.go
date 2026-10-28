package main

import postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"

const (
	databaseName          = "harpy"
	pgTrgmControlHash     = "17c2745da2a706e1443ee363ab7424ced871a9a652e86bf55d25d139f43ee27b"
	vectorControlHash     = "db4ad0045e6622228471acdcb8fc9aeeec0cd6cf9666569d26f723751ea95fcb"
	extensionManifestHash = "4ffeeb750859e5da63faef81709819c76e4c1ff5d9064e44658d9826835c42e7"
)

func extensionManifest() postgresv1.ExtensionManifest {
	return postgresv1.ExtensionManifest{
		ContractVersion: "postgres-extension-manifest.v1", DatabaseOrigin: "fresh",
		PostgresImageDigest:   postgresv1.PostgresImageDigest,
		PostgresEngineVersion: postgresv1.PostgresEngineVersion,
		ControlPathPolicy: postgresv1.ControlPathPolicy{
			Resolver: "pg_config_sharedir", RelativeDirectory: "extension",
		},
		Extensions: []postgresv1.PostgresExtension{
			{Name: "pg_trgm", Version: "1.6", Install: true, Schema: "platform_ext",
				Owner: "harpy_extension_admin", ControlFile: "pg_trgm.control",
				ControlFileSHA256: pgTrgmControlHash,
				Capabilities:      []string{"trigram_search"}},
			{Name: "vector", Version: "0.8.3", Install: true, Schema: "platform_ext",
				Owner: "harpy_extension_admin", ControlFile: "vector.control",
				ControlFileSHA256: vectorControlHash,
				Capabilities:      []string{"vector_similarity"}},
		},
	}
}

func rolesManifest() postgresv1.RolesManifest {
	names := []string{
		"harpy_owner", "harpy_cutover_authority_owner", "harpy_runtime",
		"forge_runtime",
		"harpy_forms_record_reader", "harpy_forms_record_writer",
		"harpy_v2_runtime", "harpy_migrator", "harpy_cluster_admin",
		"harpy_extension_admin", "harpy_adopter", "harpy_cutover_decoder",
		"harpy_cutover_finalizer", "harpy_backup_reader",
		"harpy_backup", "harpy_restore",
		"harpy_reactivation_observer",
		"harpy_tenant_migration_executor",
	}
	roles := make([]postgresv1.PostgresRole, 0, len(names))
	for _, name := range names {
		role := postgresv1.PostgresRole{
			Name: name, ConnectionLimit: -1, PasswordState: "null", ValidUntil: "infinity",
		}
		switch name {
		case "harpy_v2_runtime", "harpy_migrator",
			"harpy_backup", "harpy_restore", "harpy_reactivation_observer":
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

func privilegesManifest() postgresv1.PrivilegesManifest {
	memberships := []postgresv1.RoleMembership{
		{Member: "harpy_v2_runtime", Role: "harpy_runtime", SetOption: true},
		{Member: "harpy_v2_runtime", Role: "harpy_forms_record_reader", SetOption: true},
		{Member: "harpy_v2_runtime", Role: "harpy_forms_record_writer", SetOption: true},
		{Member: "harpy_migrator", Role: "harpy_owner", SetOption: true},
		{Member: "harpy_backup", Role: "harpy_backup_reader", SetOption: true},
		{Member: "harpy_restore", Role: "harpy_owner", SetOption: true},
	}
	grantees := []string{
		"harpy_backup", "harpy_cluster_admin", "harpy_extension_admin",
		"harpy_migrator", "harpy_reactivation_observer",
		"harpy_restore", "harpy_runtime",
		"harpy_v2_runtime",
	}
	grants := make([]postgresv1.ACLGrant, 0, len(grantees))
	for _, grantee := range grantees {
		grants = append(grants, postgresv1.ACLGrant{
			Object: databaseObject(), Grantor: "harpy_owner", Grantee: grantee,
			Privileges: []string{"CONNECT"},
		})
	}
	return postgresv1.PrivilegesManifest{
		ContractVersion: "postgres-privileges-manifest.v1", DatabaseOrigin: "fresh",
		DatabaseName: databaseName, PublicDatabasePrivileges: []string{},
		PublicSchemaCreate: false, Memberships: memberships, Grants: grants,
		DefaultPrivileges: []postgresv1.DefaultPrivilege{},
		ExecutorDecisions: postgresv1.ExecutorDecisionSet{
			BootstrapRole: "harpy_cluster_admin", ExtensionRole: "harpy_extension_admin",
			RuntimeSessionRole: "harpy_v2_runtime", RuntimeCurrentRole: "harpy_runtime",
			MigratorSessionRole: "harpy_migrator", MigratorCurrentRole: "harpy_owner",
			OwnershipExecutor: "cluster_bootstrap",
		},
	}
}

func ownershipManifest(manifestID string) postgresv1.OwnershipManifest {
	return postgresv1.OwnershipManifest{
		ContractVersion: "postgres-ownership-manifest.v1", ManifestID: manifestID,
		DatabaseOrigin: "fresh", DatabaseName: databaseName,
		OwnershipExecutor: "cluster_bootstrap", Format: "inline", ObjectCount: 1,
		Objects: []postgresv1.OwnershipObject{{
			Object: databaseObject(), Owner: "harpy_owner",
			Classification: "first_party", Executor: "cluster_bootstrap",
		}},
	}
}

func databaseObject() postgresv1.CatalogObject {
	return postgresv1.CatalogObject{
		Kind: "database", DatabaseName: databaseName, ObjectName: databaseName,
	}
}
