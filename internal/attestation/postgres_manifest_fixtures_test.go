package attestation

import (
	"encoding/json"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func testMigrationManifest() postgresv1.MigrationManifest {
	return postgresv1.MigrationManifest{
		ContractVersion: "postgres-migration-manifest.v1", Engine: "goose.v3",
		PostgresMajor: 18, HistoryTable: "platform.platform_migration_history",
		Entries: []postgresv1.MigrationEntry{{
			Owner: "tenancy", Scope: "platform", Version: 20260714010000,
			TransactionMode: "transactional", Phase: "foundation",
			TargetDatabaseGeneration: 1, NextMinimumRuntimeGeneration: 1,
			Variants: postgresv1.MigrationVariants{
				Fresh: &postgresv1.MigrationVariant{
					Path: "migrations/20260714010000_foundation.fresh.sql", SHA256: testHash("a"),
				},
				V1Adopted: &postgresv1.MigrationVariant{
					Path: "migrations/20260714010000_foundation.v1_adopted.sql", SHA256: testHash("b"),
				},
			},
		}},
	}
}

func testExtensionManifest() postgresv1.ExtensionManifest {
	return postgresv1.ExtensionManifest{
		ContractVersion: "postgres-extension-manifest.v1", DatabaseOrigin: "fresh",
		PostgresImageDigest:   postgresv1.PostgresImageDigest,
		PostgresEngineVersion: postgresv1.PostgresEngineVersion,
		ControlPathPolicy: postgresv1.ControlPathPolicy{
			Resolver: "pg_config_sharedir", RelativeDirectory: "extension",
		},
		Extensions: []postgresv1.PostgresExtension{
			{
				Name: "pg_trgm", Version: "1.6", Install: true, Schema: "platform_ext",
				Owner: "harpy_extension_admin", ControlFile: "pg_trgm.control",
				ControlFileSHA256: testHash("c"), Capabilities: []string{"trigram_search"},
			},
			{
				Name: "vector", Version: "0.8.3", Install: true, Schema: "platform_ext",
				Owner: "harpy_extension_admin", ControlFile: "vector.control",
				ControlFileSHA256: testHash("d"), Capabilities: []string{"vector_similarity"},
			},
		},
	}
}

func testRolesManifest() postgresv1.RolesManifest {
	names := []string{
		"harpy_owner", "harpy_cutover_authority_owner",
		"harpy_runtime", "forge_runtime", "harpy_forms_record_reader",
		"harpy_forms_record_writer", "harpy_v2_runtime",
		"harpy_migrator", "harpy_cluster_admin", "harpy_extension_admin",
		"harpy_adopter", "harpy_cutover_decoder", "harpy_cutover_finalizer",
		"harpy_backup_reader", "harpy_backup", "harpy_restore",
		"harpy_reactivation_observer", "harpy_tenant_migration_executor",
	}
	roles := make([]postgresv1.PostgresRole, 0, len(names))
	for _, name := range names {
		roles = append(roles, testRole(name))
	}
	return postgresv1.RolesManifest{
		ContractVersion: "postgres-roles-manifest.v1", DatabaseOrigin: "fresh", Roles: roles,
	}
}

func testRole(name string) postgresv1.PostgresRole {
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
	return role
}

func testPrivilegesManifest() postgresv1.PrivilegesManifest {
	database := postgresv1.CatalogObject{
		Kind: "database", DatabaseName: "harpy", ObjectName: "harpy",
	}
	return postgresv1.PrivilegesManifest{
		ContractVersion: "postgres-privileges-manifest.v1",
		DatabaseOrigin:  "fresh", DatabaseName: "harpy",
		PublicDatabasePrivileges: []string{},
		Memberships: []postgresv1.RoleMembership{
			{Member: "harpy_v2_runtime", Role: "harpy_runtime", SetOption: true},
			{Member: "harpy_v2_runtime", Role: "harpy_forms_record_reader", SetOption: true},
			{Member: "harpy_v2_runtime", Role: "harpy_forms_record_writer", SetOption: true},
			{Member: "harpy_migrator", Role: "harpy_owner", SetOption: true},
			{Member: "harpy_backup", Role: "harpy_backup_reader", SetOption: true},
			{Member: "harpy_restore", Role: "harpy_owner", SetOption: true},
		},
		Grants: []postgresv1.ACLGrant{
			testDatabaseConnectGrant(database, "harpy_backup"),
			testDatabaseConnectGrant(database, "harpy_cluster_admin"),
			testDatabaseConnectGrant(database, "harpy_extension_admin"),
			testDatabaseConnectGrant(database, "harpy_migrator"),
			testDatabaseConnectGrant(database, "harpy_reactivation_observer"),
			testDatabaseConnectGrant(database, "harpy_restore"),
			testDatabaseConnectGrant(database, "harpy_runtime"),
			testDatabaseConnectGrant(database, "harpy_v2_runtime"),
		},
		DefaultPrivileges: []postgresv1.DefaultPrivilege{{
			Owner: "harpy_owner", SchemaName: "platform", ObjectKind: "table",
			Grantee: "harpy_runtime", Privileges: []string{"SELECT", "INSERT", "UPDATE", "DELETE"},
		}},
		ExecutorDecisions: postgresv1.ExecutorDecisionSet{
			BootstrapRole: "harpy_cluster_admin", ExtensionRole: "harpy_extension_admin",
			RuntimeSessionRole: "harpy_v2_runtime", RuntimeCurrentRole: "harpy_runtime",
			MigratorSessionRole: "harpy_migrator", MigratorCurrentRole: "harpy_owner",
			OwnershipExecutor: "cluster_bootstrap",
		},
	}
}

func testDatabaseConnectGrant(
	database postgresv1.CatalogObject,
	grantee string,
) postgresv1.ACLGrant {
	return postgresv1.ACLGrant{
		Object: database, Grantor: "harpy_owner", Grantee: grantee,
		Privileges: []string{"CONNECT"},
	}
}

func testOwnershipManifest() postgresv1.OwnershipManifest {
	return postgresv1.OwnershipManifest{
		ContractVersion: "postgres-ownership-manifest.v1",
		ManifestID:      "11111111-1111-4111-8111-111111111111",
		DatabaseOrigin:  "fresh", DatabaseName: "harpy",
		OwnershipExecutor: "cluster_bootstrap", Format: "inline", ObjectCount: 1,
		Objects: []postgresv1.OwnershipObject{testOwnershipObject()},
	}
}

func testOwnershipObject() postgresv1.OwnershipObject {
	return postgresv1.OwnershipObject{
		Object: postgresv1.CatalogObject{
			Kind: "database", DatabaseName: "harpy", ObjectName: "harpy",
		},
		Owner: "harpy_owner", Classification: "first_party", Executor: "cluster_bootstrap",
	}
}

func testDocument(t *testing.T, value any) []byte {
	t.Helper()
	document, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func testHash(character string) string { return strings.Repeat(character, 64) }
