package billing

import (
	"context"

	campaignapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/campaign"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const campaignRetireOperation = "billing.catalog.campaign.retire"

func (runtime *GlobalHTTPRuntime) RetireCampaign(
	ctx context.Context, input billingapi.RetireCampaignCommand,
) (billingapi.CampaignView, error) {
	command, readback, err := commandCapabilities(input.Access, campaignRetireOperation)
	if err != nil {
		return billingapi.CampaignView{}, err
	}
	identifier, idErr := parseGlobalHTTPV7(input.CampaignID)
	if idErr != nil {
		return billingapi.CampaignView{}, billingapi.ErrInvalid
	}
	var actionErr error
	err = runtime.catalog.Command(ctx, command,
		func(scoped context.Context, scope platformtx.GlobalScope, _ platformtx.OwnerAuthority) error {
			store, bindErr := runtime.campaignStore(scope)
			commands, buildErr := campaignapp.NewCommands(store, runtime.nextID, runtime.clock)
			if bindErr != nil || buildErr != nil {
				actionErr = billingapi.ErrAuthorityUnavailable
				return actionErr
			}
			_, actionErr = commands.Retire(scoped, campaignapp.RetireRequest{
				ActorID: input.Access.PrincipalID(), IdempotencyKey: input.IdempotencyKey,
				CampaignID: identifier, ExpectedRevision: input.ExpectedRevision,
				Reason: input.Reason,
			})
			return actionErr
		})
	if actionErr != nil {
		return billingapi.CampaignView{}, normalizeCommandError(actionErr)
	}
	if err != nil {
		return billingapi.CampaignView{}, billingapi.ErrAuthorityUnavailable
	}
	return normalizeCommandReadback(runtime.readCampaign(ctx, readback, identifier))
}
