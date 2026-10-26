package billing

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func (handler *Handler) TenantRoutes() []httpedge.Route {
	if handler == nil {
		return nil
	}
	return []httpedge.Route{
		{Method: http.MethodGet, Pattern: "/api/billing/v2/subscriptions/current", Handler: http.HandlerFunc(handler.getCurrentSubscription)},
		{Method: http.MethodGet, Pattern: "/api/billing/v2/usage", Handler: http.HandlerFunc(handler.getUsage)},
		{Method: http.MethodPost, Pattern: "/api/billing/v2/subscriptions/current/changes", Handler: http.HandlerFunc(handler.scheduleChange)},
		{Method: http.MethodPost, Pattern: "/api/billing/v2/subscriptions/current/cancellations", Handler: http.HandlerFunc(handler.scheduleCancellation)},
		{Method: http.MethodPost, Pattern: "/api/billing/v2/subscriptions/current/cancellations/resume", Handler: http.HandlerFunc(handler.resumeCancellation)},
		{Method: http.MethodPost, Pattern: "/api/billing/v2/subscriptions/current/discount-redemptions", Handler: http.HandlerFunc(handler.redeemDiscount)},
		{Method: http.MethodPost, Pattern: "/api/billing/v2/discounts/validate", Handler: http.HandlerFunc(handler.validateDiscount)},
	}
}
