package billing

import (
	"context"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discountcatalog"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const discountCreateOperation = "billing.catalog.discount.create"

func (runtime *GlobalHTTPRuntime) CreateDiscount(
	ctx context.Context, input billingapi.CreateDiscountCommand,
) (billingapi.DiscountView, error) {
	command, readback, err := commandCapabilities(input.Access, discountCreateOperation)
	if err != nil {
		return billingapi.DiscountView{}, err
	}
	policyInput, inputErr := globalHTTPDiscountInput(input.InitialVersion)
	if policyInput.Name == "" {
		policyInput.Name = policyInput.Code
	}
	policy, policyErr := discountapp.BuildPolicy(policyInput)
	presentation, presentationErr := discountapp.BuildPresentation(policyInput)
	campaign, campaignErr := globalHTTPCampaignUpdate(input.InitialVersion.Campaign)
	if inputErr != nil || policyErr != nil || presentationErr != nil || campaignErr != nil {
		return billingapi.DiscountView{}, billingapi.ErrInvalid
	}
	var identifier uuid.UUID
	var actionErr error
	err = runtime.catalog.Command(ctx, command,
		func(scoped context.Context, scope platformtx.GlobalScope, _ platformtx.OwnerAuthority) error {
			store, bindErr := runtime.discountStore(scope)
			commands, buildErr := discountapp.NewCommands(store, runtime.nextID, runtime.clock)
			if bindErr != nil || buildErr != nil {
				actionErr = billingapi.ErrAuthorityUnavailable
				return actionErr
			}
			result, executeErr := commands.Create(scoped, discountapp.CreateRequest{
				ActorID: input.Access.PrincipalID(), IdempotencyKey: input.IdempotencyKey,
				Policy: policy, Presentation: presentation,
				Available: globalHTTPAvailable(input.InitialVersion.Available),
				Campaign:  campaign,
			})
			actionErr, identifier = executeErr, result.DiscountID
			return actionErr
		})
	if actionErr != nil {
		return billingapi.DiscountView{}, normalizeCommandError(actionErr)
	}
	if err != nil || identifier == uuid.Nil {
		return billingapi.DiscountView{}, billingapi.ErrAuthorityUnavailable
	}
	return normalizeCommandReadback(runtime.readDiscount(ctx, readback, identifier))
}
