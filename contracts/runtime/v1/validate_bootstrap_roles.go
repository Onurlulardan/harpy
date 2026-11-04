package runtimev1

import "fmt"

func validateBootstrapAuthorityRoles(
	roles []BootstrapAuthorityRole,
	_ DatabaseOrigin,
	_ BootstrapAuthorityDatabase,
) error {
	if len(roles) != 4 {
		return fmt.Errorf("bootstrap authority must contain exactly four roles")
	}
	expected := []BootstrapAuthorityRole{
		roleContract("harpy_owner", false, false, -1),
		roleContract("harpy_goose", true, false, 4),
		roleContract("harpy_api", true, true, 20),
		roleContract("harpy_worker", true, true, 20),
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
	login bool,
	inherit bool,
	connectionLimit int,
) BootstrapAuthorityRole {
	return BootstrapAuthorityRole{
		Name: name, Enabled: true, CredentialMode: "permanent",
		Login: login, Inherit: inherit, Superuser: false,
		CreateDB: false, CreateRole: false,
		Replication: false, BypassRLS: false,
		ConnectionLimit: connectionLimit,
	}
}
