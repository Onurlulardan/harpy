package billing

import (
	"context"

	campaignapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/campaign"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const campaignPublishOperation = "billing.catalog.campaign.publish"

func (runtime *GlobalHTTPRuntime) CreateCampaignVersion(
	ctx context.Context, input billingapi.CreateCampaignVersionCommand,
) (billingapi.CampaignView, error) {
	command, readback, err := commandCapabilities(input.Access, campaignPublishOperation)
	if err != nil {
		return billingapi.CampaignView{}, err
	}
	identifier, idErr := parseGlobalHTTPV7(input.CampaignID)
	publicationInput, inputErr := globalHTTPCampaignInput(input.Version)
	publication, publicationErr := campaignapp.BuildPublication(publicationInput)
	if idErr != nil || inputErr != nil || publicationErr != nil {
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
			_, actionErr = commands.Publish(scoped, campaignapp.CreateVersionRequest{
				ActorID: input.Access.PrincipalID(), IdempotencyKey: input.IdempotencyKey,
				CampaignID: identifier, ExpectedRevision: input.ExpectedRevision,
				Publication: publication,
				Available:   globalHTTPAvailable(input.Version.Available),
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
