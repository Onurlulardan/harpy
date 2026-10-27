//go:build integration

package migrator

import (
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func builtEntrypointAuthorityRoles() []runtimev1.BootstrapAuthorityRole {
	return []runtimev1.BootstrapAuthorityRole{
		builtEntrypointAuthorityRole(
			"harpy_owner", true, "permanent", false, false, false, false, -1,
		),
		builtEntrypointAuthorityRole(
			"harpy_cutover_authority_owner", true, "externally_provisioned",
			false, false, false, false, -1,
		),
		builtEntrypointAuthorityRole(
			"harpy_runtime", true, "permanent", false, false, false, false, -1,
		),
		builtEntrypointAuthorityRole(
			"forge_runtime", true, "compatibility", false, false, false, false, -1,
		),
		builtEntrypointAuthorityRole(
			"harpy_forms_record_reader", true, "permanent",
			false, false, false, false, -1,
		),
		builtEntrypointAuthorityRole(
			"harpy_forms_record_writer", true, "permanent",
			false, false, false, false, -1,
		),
		builtEntrypointAuthorityRole(
			"harpy_v2_runtime", true, "permanent", true, false, false, false, -1,
		),
		builtEntrypointAuthorityRole(
			"harpy_migrator", true, "permanent", true, false, false, false, -1,
		),
		builtEntrypointAuthorityRole(
			"harpy_v1_runtime", false, "coexistence", false, false, false, false, -1,
		),
		builtEntrypointAuthorityRole(
			"harpy_cluster_admin", true, "leased", false, false, true, true, 1,
		),
		builtEntrypointAuthorityRole(
			"harpy_extension_admin", true, "leased", false, true, false, false, 1,
		),
		builtEntrypointAuthorityRole(
			"harpy_legacy_extension_owner", false, "adopted_only",
			false, false, false, false, -1,
		),
		builtEntrypointAuthorityRole(
			"harpy_adopter", true, "cutover_leased", false, false, false, false, 0,
		),
		builtEntrypointAuthorityRole(
			"harpy_cutover_decoder", true, "cutover_leased",
			false, false, false, false, 0,
		),
		builtEntrypointAuthorityRole(
			"harpy_cutover_finalizer", true, "cutover_finalization_leased",
			false, false, false, false, 0,
		),
		builtEntrypointAuthorityRole(
			"harpy_backup_reader", true, "permanent", false, false, false, false, -1,
		),
		builtEntrypointAuthorityRole(
			"harpy_backup", true, "permanent", true, false, false, false, -1,
		),
		builtEntrypointAuthorityRole(
			"harpy_restore", true, "permanent", true, false, false, false, -1,
		),
		builtEntrypointAuthorityRole(
			"harpy_reactivation_observer", true, "permanent",
			true, false, false, false, -1,
		),
	}
}

func builtEntrypointAuthorityRole(
	name string,
	enabled bool,
	credentialMode string,
	login, superuser, createDB, createRole bool,
	connectionLimit int,
) runtimev1.BootstrapAuthorityRole {
	return runtimev1.BootstrapAuthorityRole{
		Name: name, Enabled: enabled, CredentialMode: credentialMode, Login: login,
		Superuser: superuser, CreateDB: createDB, CreateRole: createRole,
		ConnectionLimit: connectionLimit,
	}
}

func builtEntrypointAuthorityMembership(
	member, role string,
	enabled, inherit, set bool,
) runtimev1.BootstrapAuthorityMembership {
	return runtimev1.BootstrapAuthorityMembership{
		Member: member, Role: role, Enabled: enabled, Inherit: inherit, Set: set,
		Grantor: "harpy_cluster_admin",
	}
}

func builtEntrypointAuthorityLease(
	id, principal string,
	notBefore, expiresAt time.Time,
) runtimev1.BootstrapCredentialLease {
	return runtimev1.BootstrapCredentialLease{
		LeaseID: id, Principal: principal,
		NotBefore: notBefore.Format(time.RFC3339), ExpiresAt: expiresAt.Format(time.RFC3339),
		MaxDurationSeconds: 900, MaxConnections: 1,
		CleanupPolicy: "disable_login_clear_password_terminate_sessions",
	}
}
