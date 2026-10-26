package sessioncapability

import "strings"

var TenantReleaseCatalog = []string{
	"catalog.read",
	"catalog.install", "catalog.upgrade", "catalog.disable", "catalog.enable",
	"catalog.restore", "catalog.retire", "catalog.purge",
	"catalog.demo.manage", "catalog.menu.manage",
	"analytics.usage.read",
	"bug_reports.read", "bug_reports.manage",
	"billing.subscription.read", "billing.subscription.manage",
	"billing.usage.read", "billing.entitlements.read",
}

var tenantReleaseCapabilities = func() map[string]struct{} {
	result := make(map[string]struct{}, len(TenantReleaseCatalog))
	for _, capability := range TenantReleaseCatalog {
		result[capability] = struct{}{}
	}
	return result
}()

func TenantReleaseCapabilityKnown(capability string) bool {
	_, known := tenantReleaseCapabilities[capability]
	return known
}

func ValidTenantCapability(capability string) bool {
	return validCapability(capability) && !strings.HasPrefix(capability, "platform.")
}

func validCapability(capability string) bool {
	if len(capability) < 3 || len(capability) > 128 ||
		capability[0] < 'a' || capability[0] > 'z' {
		return false
	}
	for _, character := range capability[1:] {
		if (character < 'a' || character > 'z') &&
			(character < '0' || character > '9') &&
			character != '.' && character != '_' && character != '-' &&
			character != ':' {
			return false
		}
	}
	return !strings.Contains(capability, "..") && capability[len(capability)-1] != '.'
}
