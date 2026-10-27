package billing

import (
	"context"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/google/uuid"
)

const campaignReadOperation = "billing.catalog.campaign.read"

func (runtime *GlobalHTTPRuntime) GetCampaign(
	ctx context.Context, access billingapi.Access, rawID string,
) (billingapi.CampaignView, error) {
	identifier, err := uuid.Parse(rawID)
	if err != nil {
		return billingapi.CampaignView{}, billingapi.ErrInvalid
	}
	capability, err := catalogReadbackCapability(access, campaignReadOperation)
	if err != nil {
		return billingapi.CampaignView{}, err
	}
	return runtime.readCampaign(ctx, capability, identifier)
}
