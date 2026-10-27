package billing

import (
	campaignapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/campaign"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func globalHTTPCampaignInput(
	input billingapi.CampaignVersionInput,
) (campaignapp.PublicationInput, error) {
	priority, priorityErr := int32Value(input.Priority)
	start, startErr := parseTimestamp(input.StartsAt)
	end, endErr := parseOptionalTimestamp(input.EndsAt)
	if priorityErr != nil || startErr != nil || endErr != nil {
		return campaignapp.PublicationInput{}, errInvalidInput
	}
	return campaignapp.PublicationInput{
		Name: input.Name, Description: input.Description,
		BannerText: input.BannerText, AutoApply: input.AutoApply,
		Priority: priority,
		StartsAt: start, EndsAt: end,
		DiscountVersionIDs: append([]string(nil), input.DiscountVersionIDs...),
	}, nil
}

func globalHTTPAvailable(value *bool) bool {
	return value == nil || *value
}
