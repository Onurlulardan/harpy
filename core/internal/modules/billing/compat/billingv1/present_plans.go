package billingv1

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) presentAdminPlans(
	ctx context.Context,
	access billingapi.Access,
	offerings []billingapi.OfferingView,
) ([]planWire, error) {
	selected, err := selectAdminPlans(offerings)
	if err != nil {
		return nil, err
	}
	result := make([]planWire, len(selected))
	for index, offering := range selected {
		identity, resolveErr := handler.resolvePlanIdentity(ctx, access, offering.ID)
		if resolveErr != nil {
			return nil, resolveErr
		}
		result[index], err = planWithLegacyID(offering, identity)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (handler *Handler) presentPublicPlans(
	ctx context.Context,
	offerings []billingapi.OfferingView,
) ([]planWire, error) {
	selected, err := selectPublicPlans(offerings)
	if err != nil {
		return nil, err
	}
	result := make([]planWire, len(selected))
	for index, offering := range selected {
		identity, resolveErr := handler.resolvePublicPlanIdentity(ctx, offering.ID)
		if resolveErr != nil {
			return nil, resolveErr
		}
		result[index], err = planWithLegacyID(offering, identity)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
