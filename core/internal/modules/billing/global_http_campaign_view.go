package billing

import (
	catalogquery "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func campaignView(snapshot catalogquery.CampaignSnapshot) billingapi.CampaignView {
	discounts := snapshot.Discounts()
	discountViews := make([]billingapi.DiscountView, len(discounts))
	for index, discount := range discounts {
		discountViews[index] = globalHTTPDiscountView(discount)
	}
	return billingapi.CampaignView{
		ID: snapshot.ID().String(), LegacyID: snapshot.LegacyID().String(),
		CreatedBy: snapshot.CreatedBy().String(), Key: snapshot.Key(),
		Revision:  snapshot.Revision(),
		VersionID: snapshot.VersionID().String(), Version: snapshot.Version(),
		State: snapshot.State(), Name: snapshot.Name(), Description: snapshot.Description(),
		BannerText: snapshot.BannerText(), AutoApply: snapshot.AutoApply(),
		Available: snapshot.Available(),
		Priority:  int(snapshot.Priority()), StartsAt: formatTimestamp(snapshot.StartsAt()),
		EndsAt: formatOptionalTimestamp(snapshot.EndsAt()), Discounts: discountViews,
		CreatedAt: formatTimestamp(snapshot.CreatedAt()),
		UpdatedAt: formatTimestamp(snapshot.UpdatedAt()),
	}
}
