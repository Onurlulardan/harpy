package billing

import (
	"context"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discountcatalog"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const discountRetireOperation = "billing.catalog.discount.retire"

func (runtime *GlobalHTTPRuntime) RetireDiscount(
	ctx context.Context, input billingapi.RetireDiscountCommand,
) (billingapi.DiscountView, error) {
	command, readback, err := commandCapabilities(input.Access, discountRetireOperation)
	if err != nil {
		return billingapi.DiscountView{}, err
	}
	identifier, idErr := parseGlobalHTTPV7(input.DiscountID)
	if idErr != nil {
		return billingapi.DiscountView{}, billingapi.ErrInvalid
	}
	var actionErr error
	err = runtime.catalog.Command(ctx, command,
		func(scoped context.Context, scope platformtx.GlobalScope, _ platformtx.OwnerAuthority) error {
			store, bindErr := runtime.discountStore(scope)
			commands, buildErr := discountapp.NewCommands(store, runtime.nextID, runtime.clock)
			if bindErr != nil || buildErr != nil {
				actionErr = billingapi.ErrAuthorityUnavailable
				return actionErr
			}
			_, actionErr = commands.Retire(scoped, discountapp.RetireRequest{
				ActorID: input.Access.PrincipalID(), IdempotencyKey: input.IdempotencyKey,
				DiscountID: identifier, ExpectedRevision: input.ExpectedRevision,
				Reason: input.Reason,
			})
			return actionErr
		})
	if actionErr != nil {
		return billingapi.DiscountView{}, normalizeCommandError(actionErr)
	}
	if err != nil {
		return billingapi.DiscountView{}, billingapi.ErrAuthorityUnavailable
	}
	return normalizeCommandReadback(runtime.readDiscount(ctx, readback, identifier))
}
