package billing

import (
	"context"
	"errors"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discount"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *tenantOperations) RedeemDiscount(
	ctx context.Context,
	request billingapi.RedeemDiscountCommand,
) (billingapi.SubscriptionView, error) {
	authority, err := application.resolve(
		ctx, request.Access, billingapi.PermissionBillingManage,
	)
	if err != nil {
		return billingapi.SubscriptionView{}, err
	}
	code, err := discountapp.CanonicalCode(request.Code)
	if err != nil || !validIdempotencyKey(request.IdempotencyKey) {
		return billingapi.SubscriptionView{}, billingapi.ErrInvalid
	}
	eligibility, err := application.readDiscount(
		ctx, request.Access, authority, code, uuid.Nil,
	)
	if err != nil {
		if errors.Is(err, billingapi.ErrNotFound) {
			return billingapi.SubscriptionView{}, billingapi.ErrConflict
		}
		return billingapi.SubscriptionView{}, err
	}
	if request.SubscriptionID != "" {
		identifier, valid := parseV7(request.SubscriptionID)
		if !valid {
			return billingapi.SubscriptionView{}, billingapi.ErrInvalid
		}
		if identifier != eligibility.SubscriptionID {
			return billingapi.SubscriptionView{}, billingapi.ErrNotFound
		}
	}
	sourceID, err := operationIdentity(
		request.Access, discountapp.OperationRedeem, request.IdempotencyKey,
	)
	if err != nil {
		return billingapi.SubscriptionView{}, err
	}
	err = application.inCommand(
		ctx, request.Access, authority, discountapp.OperationRedeem,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			commands, bindErr := application.bindings.DiscountCommands(scope)
			if bindErr != nil || commands == nil {
				return billingapi.ErrAuthorityUnavailable
			}
			commandErr := commands.Redeem(scoped, billingapi.TenantDiscountRedemption{
				Code: code, SubscriptionID: eligibility.SubscriptionID,
				OfferingVersionID: eligibility.OfferingVersionID,
				PriceVersionID:    eligibility.PriceVersionID, Currency: eligibility.Currency,
				SourceOperationID: sourceID,
			})
			return commandErr
		},
	)
	if err != nil {
		return billingapi.SubscriptionView{}, normalizeDiscountCommandError(err)
	}
	current, err := application.readSubscription(ctx, request.Access, authority)
	if err != nil {
		return billingapi.SubscriptionView{}, err
	}
	return current.View, nil
}
