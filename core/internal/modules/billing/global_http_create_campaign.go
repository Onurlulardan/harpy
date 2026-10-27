package billing

import (
	"context"

	campaignapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/campaign"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const campaignCreateOperation = "billing.catalog.campaign.create"

func (runtime *GlobalHTTPRuntime) CreateCampaign(
	ctx context.Context, input billingapi.CreateCampaignCommand,
) (billingapi.CampaignView, error) {
	command, readback, err := commandCapabilities(input.Access, campaignCreateOperation)
	if err != nil {
		return billingapi.CampaignView{}, err
	}
	key, keyErr := campaignapp.ParseKey(input.Key)
	publicationInput, inputErr := globalHTTPCampaignInput(input.InitialVersion)
	publication, publicationErr := campaignapp.BuildPublication(publicationInput)
	if keyErr != nil || inputErr != nil || publicationErr != nil {
		return billingapi.CampaignView{}, billingapi.ErrInvalid
	}
	var identifier uuid.UUID
	var actionErr error
	err = runtime.catalog.Command(ctx, command,
		func(scoped context.Context, scope platformtx.GlobalScope, _ platformtx.OwnerAuthority) error {
			store, bindErr := runtime.campaignStore(scope)
			commands, buildErr := campaignapp.NewCommands(store, runtime.nextID, runtime.clock)
			if bindErr != nil || buildErr != nil {
				actionErr = billingapi.ErrAuthorityUnavailable
				return actionErr
			}
			result, executeErr := commands.Create(scoped, campaignapp.CreateRequest{
				ActorID: input.Access.PrincipalID(), IdempotencyKey: input.IdempotencyKey,
				Key: key, Publication: publication,
				Available: globalHTTPAvailable(input.InitialVersion.Available),
			})
			actionErr, identifier = executeErr, result.CampaignID
			return actionErr
		})
	if actionErr != nil {
		return billingapi.CampaignView{}, normalizeCommandError(actionErr)
	}
	if err != nil || identifier == uuid.Nil {
		return billingapi.CampaignView{}, billingapi.ErrAuthorityUnavailable
	}
	return normalizeCommandReadback(runtime.readCampaign(ctx, readback, identifier))
}
