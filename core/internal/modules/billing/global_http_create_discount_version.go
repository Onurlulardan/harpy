package billing

import (
	"context"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discountcatalog"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const discountPublishOperation = "billing.catalog.discount.publish"

func (runtime *GlobalHTTPRuntime) CreateDiscountVersion(
	ctx context.Context, input billingapi.CreateDiscountVersionCommand,
) (billingapi.DiscountView, error) {
	command, readback, err := commandCapabilities(input.Access, discountPublishOperation)
	if err != nil {
		return billingapi.DiscountView{}, err
	}
	identifier, idErr := parseGlobalHTTPV7(input.DiscountID)
	policyInput, inputErr := globalHTTPDiscountInput(input.Version)
	if policyInput.Name == "" {
		policyInput.Name = policyInput.Code
	}
	policy, policyErr := discountapp.BuildPolicy(policyInput)
	presentation, presentationErr := discountapp.BuildPresentation(policyInput)
	campaign, campaignErr := globalHTTPCampaignUpdate(input.Version.Campaign)
	if idErr != nil || inputErr != nil || policyErr != nil ||
		presentationErr != nil || campaignErr != nil {
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
			_, actionErr = commands.Publish(scoped, discountapp.CreateVersionRequest{
				ActorID: input.Access.PrincipalID(), IdempotencyKey: input.IdempotencyKey,
				DiscountID: identifier, ExpectedRevision: input.ExpectedRevision,
				Policy: policy, Presentation: presentation,
				Available: input.Version.Available, Campaign: campaign,
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
