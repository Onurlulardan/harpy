package billingv1

import billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"

func selectAdminPlans(
	offerings []billingapi.OfferingView,
) ([]billingapi.OfferingView, error) {
	latest := make(map[string]billingapi.OfferingView, len(offerings))
	for _, offering := range offerings {
		if offering.ID == "" || offering.VersionID == "" || offering.Version < 1 {
			return nil, ErrSnapshotRequired
		}
		if offering.State == "retired" {
			continue
		}
		if offering.State != "active" {
			return nil, ErrSnapshotRequired
		}
		current, exists := latest[offering.ID]
		if !exists || offering.Version > current.Version {
			latest[offering.ID] = offering
			continue
		}
		if offering.Version == current.Version {
			return nil, ErrSnapshotRequired
		}
	}
	selected := make([]billingapi.OfferingView, 0, len(latest))
	for _, offering := range latest {
		selected = append(selected, offering)
	}
	return sortOfferings(selected), nil
}

func selectPublicPlans(
	offerings []billingapi.OfferingView,
) ([]billingapi.OfferingView, error) {
	selected, err := selectAdminPlans(offerings)
	if err != nil {
		return nil, err
	}
	for _, offering := range selected {
		if !offering.Public || !offering.Available {
			return nil, ErrSnapshotRequired
		}
	}
	return selected, nil
}
