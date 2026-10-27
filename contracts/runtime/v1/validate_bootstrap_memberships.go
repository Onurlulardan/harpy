package runtimev1

import "fmt"

func validateBootstrapAuthorityMemberships(
	memberships []BootstrapAuthorityMembership,
	origin DatabaseOrigin,
) error {
	if len(memberships) != 7 {
		return fmt.Errorf("bootstrap authority must contain exactly seven memberships")
	}
	expected := []BootstrapAuthorityMembership{
		membershipContract("harpy_v2_runtime", "harpy_runtime", true, false, true),
		membershipContract(
			"harpy_v2_runtime", "harpy_forms_record_reader", true, false, true,
		),
		membershipContract(
			"harpy_v2_runtime", "harpy_forms_record_writer", true, false, true,
		),
		membershipContract("harpy_migrator", "harpy_owner", true, false, true),
		membershipContract(
			"harpy_backup", "harpy_backup_reader", true, false, true,
		),
		membershipContract("harpy_restore", "harpy_owner", true, false, true),
		membershipContract(
			"harpy_v1_runtime", "harpy_runtime",
			origin == DatabaseOriginV1Adopted, true, false,
		),
	}
	for index := range expected {
		if memberships[index] != expected[index] {
			return fmt.Errorf("bootstrap authority membership %d is invalid", index)
		}
	}
	return nil
}

func membershipContract(
	member, role string,
	enabled, inherit, set bool,
) BootstrapAuthorityMembership {
	return BootstrapAuthorityMembership{
		Member: member, Role: role, Enabled: enabled,
		Admin: false, Inherit: inherit, Set: set,
		Grantor: "harpy_cluster_admin",
	}
}
