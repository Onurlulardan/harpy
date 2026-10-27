package billing

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func (handler *Handler) PublicRoutes() []httpedge.Route {
	if handler == nil {
		return nil
	}
	return []httpedge.Route{
		{Method: http.MethodGet, Pattern: "/api/billing/v2/offerings", Handler: http.HandlerFunc(handler.listPublicOfferings)},
		{Method: http.MethodGet, Pattern: "/api/billing/v2/promotions/active", Handler: http.HandlerFunc(handler.getActiveCampaign)},
	}
}
