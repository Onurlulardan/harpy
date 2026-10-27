package menu

import menudomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/menu"

const TenantAuthoredPackageID = "018f6000-0000-7000-8000-000000000099"

func activateTenantAuthored(active map[string]bool) map[string]bool {
	if active == nil {
		active = map[string]bool{}
	}
	active[TenantAuthoredPackageID] = true
	return active
}

func allowTenantAuthoredTargets(items []menudomain.Item, allowed map[string]bool) {
	for _, item := range items {
		if item.PackageID.String() != TenantAuthoredPackageID || item.TargetID == nil {
			continue
		}
		allowed[item.TargetID.String()] = true
	}
}
