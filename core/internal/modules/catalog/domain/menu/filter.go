package menu

import (
	"sort"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

type TargetAuthorizer interface {
	Allows(packagecontract.MenuKind, packagecontract.ComponentID) bool
}

func Effective(items []Item, activePackages map[string]bool, authorizer TargetAuthorizer) []Item {
	if authorizer == nil {
		return nil
	}
	result := make([]Item, 0, len(items))
	for _, item := range items {
		if !item.Visible || !activePackages[item.PackageID.String()] {
			continue
		}
		if item.Kind != packagecontract.MenuGroup && item.Kind != packagecontract.MenuExternalLink {
			if item.TargetID == nil || !authorizer.Allows(item.Kind, *item.TargetID) {
				continue
			}
		}
		result = append(result, item)
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Order != result[j].Order {
			return result[i].Order < result[j].Order
		}
		return result[i].ID.String() < result[j].ID.String()
	})
	return result
}
