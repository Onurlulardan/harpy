package postgresv1

import (
	"fmt"
	"strings"
)

func (manifest PrivilegesManifest) Validate() error {
	origin := DatabaseOrigin(manifest.DatabaseOrigin)
	if manifest.ContractVersion != "postgres-privileges-manifest.v1" || !origin.Valid() {
		return fmt.Errorf("privileges manifest header is invalid")
	}
	if err := validateDatabaseName(manifest.DatabaseName, "database_name"); err != nil {
		return err
	}
	if manifest.PublicDatabasePrivileges == nil || len(manifest.PublicDatabasePrivileges) != 0 ||
		manifest.PublicSchemaCreate {
		return fmt.Errorf("PUBLIC privilege contract is invalid")
	}
	if err := validateMemberships(manifest.Memberships, origin); err != nil {
		return err
	}
	if err := validateACLGrants(manifest.Grants, manifest.DatabaseName); err != nil {
		return err
	}
	if err := validateDirectDatabaseConnectGrants(manifest.Grants, manifest.DatabaseName); err != nil {
		return err
	}
	if err := validateDefaultPrivileges(manifest.DefaultPrivileges); err != nil {
		return err
	}
	return validateExecutorDecisions(manifest.ExecutorDecisions, origin)
}

func validateMemberships(actual []RoleMembership, origin DatabaseOrigin) error {
	expected := []RoleMembership{
		{Member: "harpy_v2_runtime", Role: "harpy_runtime", SetOption: true},
		{Member: "harpy_v2_runtime", Role: "harpy_forms_record_reader", SetOption: true},
		{Member: "harpy_v2_runtime", Role: "harpy_forms_record_writer", SetOption: true},
		{Member: "harpy_migrator", Role: "harpy_owner", SetOption: true},
		{Member: "harpy_backup", Role: "harpy_backup_reader", SetOption: true},
		{Member: "harpy_restore", Role: "harpy_owner", SetOption: true},
	}
	if origin == DatabaseOriginV1Adopted {
		expected = append(expected, RoleMembership{
			Member: "harpy_v1_runtime", Role: "harpy_runtime", InheritOption: true,
		})
	}
	if actual == nil || len(actual) != len(expected) {
		return fmt.Errorf("membership set is incomplete")
	}
	for index := range expected {
		if actual[index] != expected[index] {
			return fmt.Errorf("membership %d differs from contract", index)
		}
	}
	return nil
}

func validateACLGrants(grants []ACLGrant, databaseName string) error {
	if grants == nil || len(grants) > 100000 {
		return fmt.Errorf("ACL grant set is invalid")
	}
	keys := make([]string, 0, len(grants))
	for index, grant := range grants {
		if err := grant.Object.Validate(); err != nil {
			return fmt.Errorf("ACL grant %d: %w", index, err)
		}
		if grant.Object.DatabaseName != databaseName {
			return fmt.Errorf("ACL grant %d targets another database", index)
		}
		if err := validateIdentifier(grant.Grantor, "grantor"); err != nil {
			return err
		}
		if grant.Grantee == "PUBLIC" {
			return fmt.Errorf("ACL grant %d grants PUBLIC authority", index)
		} else {
			if err := validateIdentifier(grant.Grantee, "grantee"); err != nil {
				return err
			}
		}
		if grant.GrantOption || !validatePrivilegeList(grant.Object.Kind, grant.Privileges) {
			return fmt.Errorf("ACL grant %d privilege set is invalid", index)
		}
		keys = append(keys, aclGrantKey(grant))
	}
	return validateOrderedUnique(keys, "ACL grants")
}

func validateDefaultPrivileges(privileges []DefaultPrivilege) error {
	if privileges == nil || len(privileges) > 10000 {
		return fmt.Errorf("default privilege set is invalid")
	}
	keys := make([]string, 0, len(privileges))
	for index, privilege := range privileges {
		if privilege.Owner != "harpy_owner" || privilege.Grantee != "harpy_runtime" ||
			privilege.GrantOption || validateIdentifier(privilege.SchemaName, "schema") != nil ||
			!validatePrivilegeList(privilege.ObjectKind, privilege.Privileges) {
			return fmt.Errorf("default privilege %d is invalid", index)
		}
		keys = append(keys, strings.Join([]string{
			privilege.Owner, privilege.SchemaName, privilege.ObjectKind,
			privilege.Grantee, strings.Join(privilege.Privileges, ","),
		}, "|"))
	}
	return validateOrderedUnique(keys, "default privileges")
}

func validateExecutorDecisions(value ExecutorDecisionSet, origin DatabaseOrigin) error {
	executor := "cluster_bootstrap"
	if origin == DatabaseOriginV1Adopted {
		executor = "external_root_provisioner"
	}
	if value.BootstrapRole != "harpy_cluster_admin" ||
		value.ExtensionRole != "harpy_extension_admin" ||
		value.RuntimeSessionRole != "harpy_v2_runtime" ||
		value.RuntimeCurrentRole != "harpy_runtime" ||
		value.MigratorSessionRole != "harpy_migrator" ||
		value.MigratorCurrentRole != "harpy_owner" || value.OwnershipExecutor != executor {
		return fmt.Errorf("executor decisions are invalid")
	}
	return nil
}
