package runtimev1

import "fmt"

func validateBootstrapAuthorityRoles(
	roles []BootstrapAuthorityRole,
	origin DatabaseOrigin,
	database BootstrapAuthorityDatabase,
) error {
	if len(roles) != 16 {
		return fmt.Errorf("bootstrap authority must contain exactly sixteen roles")
	}
	v1Enabled := origin == DatabaseOriginV1Adopted
	legacyEnabled := origin == DatabaseOriginV1Adopted &&
		database.ExtensionOwnerMode == "adopted_reassigned"
	expected := []BootstrapAuthorityRole{
		roleContract("harpy_owner", true, "permanent", false, false, false, false, -1),
		roleContract("harpy_cutover_authority_owner", true, "externally_provisioned", false, false, false, false, -1),
		roleContract("harpy_runtime", true, "permanent", false, false, false, false, -1),
		roleContract("harpy_v2_runtime", true, "permanent", true, false, false, false, -1),
		roleContract("harpy_migrator", true, "permanent", true, false, false, false, -1),
		roleContract("harpy_v1_runtime", v1Enabled, "coexistence", v1Enabled, false, false, false, -1),
		roleContract("harpy_cluster_admin", true, "leased", false, false, true, true, 1),
		roleContract("harpy_extension_admin", true, "leased", false, true, false, false, 1),
		roleContract("harpy_legacy_extension_owner", legacyEnabled, "adopted_only", false, false, false, false, -1),
		roleContract("harpy_adopter", true, "cutover_leased", false, false, false, false, 0),
		roleContract("harpy_cutover_decoder", true, "cutover_leased", false, false, false, false, 0),
		roleContract("harpy_cutover_finalizer", true, "cutover_finalization_leased", false, false, false, false, 0),
		roleContract("harpy_backup_reader", true, "permanent", false, false, false, false, -1),
		roleContract("harpy_backup", true, "permanent", true, false, false, false, -1),
		roleContract("harpy_restore", true, "permanent", true, false, false, false, -1),
		roleContract("harpy_reactivation_observer", true, "permanent", true, false, false, false, -1),
	}
	for index := range expected {
		if roles[index] != expected[index] {
			return fmt.Errorf("bootstrap authority role %d is invalid", index)
		}
	}
	return nil
}

func roleContract(
	name string,
	enabled bool,
	credentialMode string,
	login bool,
	superuser bool,
	createDB bool,
	createRole bool,
	connectionLimit int,
) BootstrapAuthorityRole {
	return BootstrapAuthorityRole{
		Name: name, Enabled: enabled, CredentialMode: credentialMode,
		Login: login, Inherit: false, Superuser: superuser,
		CreateDB: createDB, CreateRole: createRole,
		Replication: false, BypassRLS: false,
		ConnectionLimit: connectionLimit,
	}
}
