package billing

import (
	"context"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discount"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type tenantDiscountQueries struct{ queries *discountapp.Queries }

func (adapter *tenantDiscountQueries) Validate(
	ctx context.Context,
	scope platformtx.TenantScope,
	request billingapi.TenantDiscountLookup,
) (billingapi.TenantDiscountSnapshot, error) {
	value, err := adapter.queries.Validate(ctx, scope, discountapp.ValidateRequest{
		Code: request.Code, OfferingVersionID: request.OfferingVersionID,
	})
	if err != nil {
		return billingapi.TenantDiscountSnapshot{}, err
	}
	return discountSnapshot(value)
}

func (adapter *tenantDiscountQueries) ResolveCurrent(
	ctx context.Context,
	scope platformtx.TenantScope,
	code string,
) (billingapi.TenantDiscountSnapshot, error) {
	value, err := adapter.queries.ResolveCurrent(ctx, scope, code)
	if err != nil {
		return billingapi.TenantDiscountSnapshot{}, err
	}
	return discountSnapshot(value)
}

type tenantDiscountCommands struct{ commands *discountapp.Commands }

func (adapter *tenantDiscountCommands) Redeem(
	ctx context.Context,
	request billingapi.TenantDiscountRedemption,
) error {
	code, err := discountapp.ParseCode(request.Code)
	if err != nil {
		return discountapp.ErrRequest
	}
	_, err = adapter.commands.Redeem(ctx, discountapp.RedeemRequest{
		Code: code, SubscriptionID: request.SubscriptionID,
		OfferingVersionID: request.OfferingVersionID,
		PriceVersionID:    request.PriceVersionID, Currency: request.Currency,
		SourceOperationID: request.SourceOperationID,
	})
	return err
}
