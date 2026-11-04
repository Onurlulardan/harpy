package runtimev1

import "fmt"

func validateBootstrapAuthorityMemberships(
	memberships []BootstrapAuthorityMembership,
	_ DatabaseOrigin,
) error {
	if len(memberships) != 1 {
		return fmt.Errorf("bootstrap authority must contain exactly one membership")
	}
	expected := membershipContract("harpy_goose", "harpy_owner", true, false, true)
	if memberships[0] != expected {
		return fmt.Errorf("bootstrap authority membership 0 is invalid")
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
		Grantor: "harpy_owner",
	}
}
