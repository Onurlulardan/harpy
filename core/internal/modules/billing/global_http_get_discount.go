package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/google/uuid"
)

const discountReadOperation = "billing.catalog.discount.read"

func (runtime *GlobalHTTPRuntime) GetDiscount(
	ctx context.Context, access billingapi.Access, rawID string,
) (billingapi.DiscountView, error) {
	identifier, err := uuid.Parse(rawID)
	if err != nil {
		return billingapi.DiscountView{}, billingapi.ErrInvalid
	}
	capability, err := catalogReadbackCapability(access, discountReadOperation)
	if err != nil {
		return billingapi.DiscountView{}, err
	}
	return runtime.readDiscount(ctx, capability, identifier)
}
