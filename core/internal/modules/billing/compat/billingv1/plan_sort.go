package billingv1

import (
	"sort"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func sortOfferings(values []billingapi.OfferingView) []billingapi.OfferingView {
	result := append([]billingapi.OfferingView(nil), values...)
	sort.SliceStable(result, func(left, right int) bool {
		if result[left].SortOrder != result[right].SortOrder {
			return result[left].SortOrder < result[right].SortOrder
		}
		if result[left].Key != result[right].Key {
			return result[left].Key < result[right].Key
		}
		return result[left].ID < result[right].ID
	})
	return result
}
