package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/google/uuid"
)

const offeringReadOperation = "billing.catalog.offering.read"

func (runtime *GlobalHTTPRuntime) GetOffering(
	ctx context.Context, access billingapi.Access, rawID string,
) (billingapi.OfferingView, error) {
	identifier, err := uuid.Parse(rawID)
	if err != nil {
		return billingapi.OfferingView{}, billingapi.ErrInvalid
	}
	capability, err := catalogReadbackCapability(access, offeringReadOperation)
	if err != nil {
		return billingapi.OfferingView{}, err
	}
	return runtime.readOffering(ctx, capability, identifier)
}
