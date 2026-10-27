//go:build integration

package postgres

import runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"

func builtCutoverAuthorityRoles() []runtimev1.BootstrapAuthorityRole {
	return []runtimev1.BootstrapAuthorityRole{
		builtCutoverAuthorityRole("harpy_owner", true, "permanent", false, false, false, false, -1),
		builtCutoverAuthorityRole("harpy_cutover_authority_owner", true,
			"externally_provisioned", false, false, false, false, -1),
		builtCutoverAuthorityRole("harpy_runtime", true, "permanent", false, false, false, false, -1),
		builtCutoverAuthorityRole("harpy_v2_runtime", true, "permanent", true, false, false, false, -1),
		builtCutoverAuthorityRole("harpy_migrator", true, "permanent", true, false, false, false, -1),
		builtCutoverAuthorityRole("harpy_v1_runtime", false, "coexistence", false, false, false, false, -1),
		builtCutoverAuthorityRole("harpy_cluster_admin", true, "leased", false, false, true, true, 1),
		builtCutoverAuthorityRole("harpy_extension_admin", true, "leased", false, true, false, false, 1),
		builtCutoverAuthorityRole("harpy_legacy_extension_owner", false,
			"adopted_only", false, false, false, false, -1),
		builtCutoverAuthorityRole("harpy_adopter", true,
			"cutover_leased", false, false, false, false, 0),
		builtCutoverAuthorityRole("harpy_cutover_decoder", true,
			"cutover_leased", false, false, false, false, 0),
		builtCutoverAuthorityRole("harpy_cutover_finalizer", true,
			"cutover_finalization_leased", false, false, false, false, 0),
		builtCutoverAuthorityRole("harpy_backup_reader", true,
			"permanent", false, false, false, false, -1),
		builtCutoverAuthorityRole("harpy_backup", true,
			"permanent", true, false, false, false, -1),
		builtCutoverAuthorityRole("harpy_restore", true,
			"permanent", true, false, false, false, -1),
		builtCutoverAuthorityRole("harpy_reactivation_observer", true,
			"permanent", true, false, false, false, -1),
	}
}

func builtCutoverAuthorityRole(
	name string, enabled bool, mode string,
	login, superuser, createDB, createRole bool, limit int,
) runtimev1.BootstrapAuthorityRole {
	return runtimev1.BootstrapAuthorityRole{
		Name: name, Enabled: enabled, CredentialMode: mode, Login: login,
		Superuser: superuser, CreateDB: createDB, CreateRole: createRole,
		ConnectionLimit: limit,
	}
}

func builtCutoverAuthorityMembership(
	member, role string, enabled, inherit, set bool,
) runtimev1.BootstrapAuthorityMembership {
	return runtimev1.BootstrapAuthorityMembership{
		Member: member, Role: role, Enabled: enabled, Inherit: inherit, Set: set,
		Grantor: "harpy_cluster_admin",
	}
}
