package billingv1

import (
	"encoding/json"
	"fmt"
	"sort"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func catalogRootIDs(items []json.RawMessage, field string) ([]string, error) {
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		var document map[string]json.RawMessage
		if err := json.Unmarshal(item, &document); err != nil {
			return nil, fmt.Errorf("decode Billing catalog query result: %w", err)
		}
		raw, exists := document[field]
		var identity string
		if !exists || json.Unmarshal(raw, &identity) != nil || identity == "" {
			return nil, ErrPaginationEvidence
		}
		if _, exists = seen[identity]; exists {
			continue
		}
		seen[identity] = struct{}{}
		result = append(result, identity)
	}
	return result, nil
}

func sortDiscountViews(
	views []billingapi.DiscountView,
) ([]billingapi.DiscountView, error) {
	for _, view := range views {
		if !canonicalWireInstant(view.CreatedAt) {
			return nil, ErrSnapshotRequired
		}
	}
	sort.SliceStable(views, func(left, right int) bool {
		return views[left].CreatedAt > views[right].CreatedAt
	})
	return views, nil
}

func sortCampaignViews(
	views []billingapi.CampaignView,
) ([]billingapi.CampaignView, error) {
	for _, view := range views {
		if !canonicalWireInstant(view.CreatedAt) {
			return nil, ErrSnapshotRequired
		}
	}
	sort.SliceStable(views, func(left, right int) bool {
		return views[left].CreatedAt > views[right].CreatedAt
	})
	return views, nil
}
