package billing

import (
	"context"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discount"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/google/uuid"
)

func (application *tenantOperations) ValidateDiscount(
	ctx context.Context,
	request billingapi.ValidateDiscountCommand,
) (billingapi.DiscountView, error) {
	code, err := discountapp.CanonicalCode(request.Code)
	offeringID, offeringValid := parseOptionalV7(request.OfferingVersionID)
	if err != nil || !offeringValid {
		return billingapi.DiscountView{}, billingapi.ErrInvalid
	}
	authority, err := application.resolve(
		ctx, request.Access, billingapi.PermissionBillingRead,
	)
	if err != nil {
		return billingapi.DiscountView{}, err
	}
	eligibility, err := application.readDiscount(
		ctx, request.Access, authority, code, offeringID,
	)
	if err != nil {
		return billingapi.DiscountView{}, err
	}
	return eligibility.View, nil
}

func parseOptionalV7(value string) (uuid.UUID, bool) {
	if value == "" {
		return uuid.Nil, true
	}
	return parseV7(value)
}
