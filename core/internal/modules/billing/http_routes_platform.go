package billing

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func (handler *Handler) PlatformRoutes() []httpedge.Route {
	if handler == nil {
		return nil
	}
	return []httpedge.Route{
		{Method: http.MethodGet, Pattern: "/api/platform/billing/offerings", Handler: http.HandlerFunc(handler.listOfferings)},
		{Method: http.MethodPost, Pattern: "/api/platform/billing/offerings", Handler: http.HandlerFunc(handler.createOffering)},
		{Method: http.MethodPost, Pattern: "/api/platform/billing/offerings/query", Handler: http.HandlerFunc(handler.queryOfferings)},
		{Method: http.MethodGet, Pattern: "/api/platform/billing/offerings/{offering_id}", Handler: http.HandlerFunc(handler.getOffering)},
		{Method: http.MethodPost, Pattern: "/api/platform/billing/offerings/{offering_id}/versions", Handler: http.HandlerFunc(handler.createOfferingVersion)},
		{Method: http.MethodPost, Pattern: "/api/platform/billing/offerings/{offering_id}/retire", Handler: http.HandlerFunc(handler.retireOffering)},
		{Method: http.MethodGet, Pattern: "/api/platform/billing/subscriptions", Handler: http.HandlerFunc(handler.listSubscriptions)},
		{Method: http.MethodPost, Pattern: "/api/platform/billing/subscriptions/query", Handler: http.HandlerFunc(handler.querySubscriptions)},
		{Method: http.MethodPost, Pattern: "/api/platform/billing/subscriptions/{subscription_id}/transitions", Handler: http.HandlerFunc(handler.transitionSubscription)},
		{Method: http.MethodGet, Pattern: "/api/platform/billing/events", Handler: http.HandlerFunc(handler.listAuditEvents)},
		{Method: http.MethodPost, Pattern: "/api/platform/billing/events/query", Handler: http.HandlerFunc(handler.queryAuditEvents)},
		{Method: http.MethodPost, Pattern: "/api/platform/billing/revenue/query", Handler: http.HandlerFunc(handler.queryRevenue)},
		{Method: http.MethodGet, Pattern: "/api/platform/billing/discounts", Handler: http.HandlerFunc(handler.listDiscounts)},
		{Method: http.MethodPost, Pattern: "/api/platform/billing/discounts", Handler: http.HandlerFunc(handler.createDiscount)},
		{Method: http.MethodPost, Pattern: "/api/platform/billing/discounts/query", Handler: http.HandlerFunc(handler.queryDiscounts)},
		{Method: http.MethodGet, Pattern: "/api/platform/billing/discounts/{discount_id}", Handler: http.HandlerFunc(handler.getDiscount)},
		{Method: http.MethodPost, Pattern: "/api/platform/billing/discounts/{discount_id}/versions", Handler: http.HandlerFunc(handler.createDiscountVersion)},
		{Method: http.MethodPost, Pattern: "/api/platform/billing/discounts/{discount_id}/retire", Handler: http.HandlerFunc(handler.retireDiscount)},
		{Method: http.MethodGet, Pattern: "/api/platform/billing/campaigns", Handler: http.HandlerFunc(handler.listCampaigns)},
		{Method: http.MethodPost, Pattern: "/api/platform/billing/campaigns", Handler: http.HandlerFunc(handler.createCampaign)},
		{Method: http.MethodPost, Pattern: "/api/platform/billing/campaigns/query", Handler: http.HandlerFunc(handler.queryCampaigns)},
		{Method: http.MethodGet, Pattern: "/api/platform/billing/campaigns/{campaign_id}", Handler: http.HandlerFunc(handler.getCampaign)},
		{Method: http.MethodPost, Pattern: "/api/platform/billing/campaigns/{campaign_id}/versions", Handler: http.HandlerFunc(handler.createCampaignVersion)},
		{Method: http.MethodPost, Pattern: "/api/platform/billing/campaigns/{campaign_id}/retire", Handler: http.HandlerFunc(handler.retireCampaign)},
	}
}
