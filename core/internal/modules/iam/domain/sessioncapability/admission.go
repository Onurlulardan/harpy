package sessioncapability

import (
	"slices"
	"strings"
)

type Sets struct {
	Mode                 DeploymentMode
	Purpose              TenantPurpose
	PlatformCapabilities []string
	TenantCapabilities   []string
}

func (sets Sets) Valid() bool {
	return sets.Mode.Valid() && sets.Purpose.ValidFor(sets.Mode) &&
		validSortedSet(sets.PlatformCapabilities, true) &&
		validSortedSet(sets.TenantCapabilities, false)
}

func (sets Sets) AllowsPlatform(capability string) bool {
	return sets.Valid() && strings.HasPrefix(capability, "platform.") &&
		ModeAllowsPlatform(sets.Mode, capability) &&
		contains(sets.PlatformCapabilities, capability)
}

func (sets Sets) AllowsTenant(capability string) bool {
	if !sets.Valid() || !ValidTenantCapability(capability) ||
		!modeAllowsTenant(sets.Mode, sets.Purpose, capability) {
		return false
	}
	return contains(sets.TenantCapabilities, capability)
}

func validSortedSet(values []string, platform bool) bool {
	if values == nil || len(values) > 4096 || !slices.IsSorted(values) {
		return false
	}
	for index, value := range values {
		if !validCapability(value) ||
			platform != strings.HasPrefix(value, "platform.") ||
			index > 0 && values[index-1] == value {
			return false
		}
	}
	return true
}

func modeAllowsTenant(
	mode DeploymentMode, purpose TenantPurpose, capability string,
) bool {
	if purpose == PurposeSaaSControl {
		return controlTenantCapability(capability)
	}
	if strings.HasPrefix(capability, "billing.subscription.") ||
		capability == "billing.usage.read" {
		return mode == ModeSaaS && purpose == PurposeSaaSCustomer
	}
	if capability == "billing.entitlements.read" {
		return mode == ModeOnPremise && purpose == PurposeOnPremisePrimary
	}
	return purpose == PurposeOnPremisePrimary || purpose == PurposeSaaSCustomer
}

func controlTenantCapability(capability string) bool {
	for _, prefix := range []string{
		"iam.", "users.", "roles.", "organizations.", "profile.", "session.",
	} {
		if strings.HasPrefix(capability, prefix) {
			return true
		}
	}
	return false
}

func contains(values []string, value string) bool {
	_, found := slices.BinarySearch(values, value)
	return found
}
