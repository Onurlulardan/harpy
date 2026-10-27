//go:build integration

package migrator

import (
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestBuiltEntrypointPrivilegesManifestValidates(t *testing.T) {
	if err := builtEntrypointPrivilegesManifest().Validate(); err != nil {
		t.Fatalf("validate built entrypoint privileges: %v", err)
	}
}

func builtEntrypointPrivilegesManifest() postgresv1.PrivilegesManifest {
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
		"harpy_migrator", "harpy_reactivation_observer", "harpy_restore",
		"harpy_runtime", "harpy_v2_runtime",
	}
	grants := make([]postgresv1.ACLGrant, 0, len(grantees))
	for _, grantee := range grantees {
		grants = append(grants, postgresv1.ACLGrant{
			Object: postgresv1.CatalogObject{
				Kind: "database", DatabaseName: builtEntrypointDatabaseName,
				ObjectName: builtEntrypointDatabaseName,
			},
			Grantor: "harpy_owner", Grantee: grantee, Privileges: []string{"CONNECT"},
		})
	}
	return postgresv1.PrivilegesManifest{
		ContractVersion: "postgres-privileges-manifest.v1", DatabaseOrigin: "fresh",
		DatabaseName:             builtEntrypointDatabaseName,
		PublicDatabasePrivileges: []string{}, PublicSchemaCreate: false,
		Memberships: memberships, Grants: grants, DefaultPrivileges: []postgresv1.DefaultPrivilege{},
		ExecutorDecisions: postgresv1.ExecutorDecisionSet{
			BootstrapRole: "harpy_cluster_admin", ExtensionRole: "harpy_extension_admin",
			RuntimeSessionRole: "harpy_v2_runtime", RuntimeCurrentRole: "harpy_runtime",
			MigratorSessionRole: "harpy_migrator", MigratorCurrentRole: "harpy_owner",
			OwnershipExecutor: "cluster_bootstrap",
		},
	}
}
