//go:build integration

package migrator

import postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"

func builtEntrypointRolesManifest() postgresv1.RolesManifest {
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
