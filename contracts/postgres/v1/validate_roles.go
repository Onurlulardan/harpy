package postgresv1

import "fmt"

var permanentRoleOrder = []string{
	"harpy_owner",
	"harpy_cutover_authority_owner",
	"harpy_runtime",
	"harpy_v2_runtime",
	"harpy_migrator",
	"harpy_cluster_admin",
	"harpy_extension_admin",
	"harpy_adopter",
	"harpy_cutover_decoder",
	"harpy_cutover_finalizer",
	"harpy_backup_reader",
	"harpy_backup",
	"harpy_restore",
	"harpy_reactivation_observer",
	"harpy_tenant_migration_executor",
}

func (manifest RolesManifest) Validate() error {
	origin := DatabaseOrigin(manifest.DatabaseOrigin)
	if manifest.ContractVersion != "postgres-roles-manifest.v1" || !origin.Valid() {
		return fmt.Errorf("roles manifest header is invalid")
	}
	expectedNames := append([]string(nil), permanentRoleOrder...)
	if origin == DatabaseOriginV1Adopted {
		expectedNames = append(
			expectedNames, "harpy_v1_runtime", "harpy_legacy_extension_owner",
		)
	}
	if manifest.Roles == nil || len(manifest.Roles) != len(expectedNames) {
		return fmt.Errorf("roles manifest has an incomplete role set")
	}
	for index, name := range expectedNames {
		expected := expectedPostgresRole(name)
		if manifest.Roles[index] != expected {
			return fmt.Errorf("role %d differs from the at-rest catalog contract", index)
		}
	}
	return nil
}

func expectedPostgresRole(name string) PostgresRole {
	role := PostgresRole{
		Name: name, ConnectionLimit: -1,
		PasswordState: "null", ValidUntil: "infinity",
	}
	switch name {
	case "harpy_v2_runtime", "harpy_migrator", "harpy_v1_runtime",
		"harpy_backup", "harpy_restore", "harpy_reactivation_observer":
		role.Login = true
	case "harpy_cluster_admin":
		role.CreateDB = true
		role.CreateRole = true
		role.ConnectionLimit = 1
	case "harpy_extension_admin":
		role.Superuser = true
		role.ConnectionLimit = 1
	case "harpy_adopter", "harpy_cutover_decoder", "harpy_cutover_finalizer":
		role.ConnectionLimit = 0
	}
	return role
}
