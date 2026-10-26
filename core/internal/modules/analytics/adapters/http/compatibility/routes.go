package compatibilityhttp

import (
	"net/http"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

const (
	legacyOpenAPIPath   = "/api/openapi/execute/{slug}"
	legacyAnalyticsPath = "/api/analytics/execute/{slug}"
)

func (handler *Handler) Routes() []httptransport.Route {
	return []httptransport.Route{
		{Method: http.MethodGet, Pattern: legacyOpenAPIPath, Handler: http.HandlerFunc(handler.Authorize)},
		{Method: http.MethodPost, Pattern: legacyOpenAPIPath, Handler: http.HandlerFunc(handler.Authorize)},
		{Method: http.MethodGet, Pattern: legacyAnalyticsPath, Handler: http.HandlerFunc(handler.Authorize)},
		{Method: http.MethodPost, Pattern: legacyAnalyticsPath, Handler: http.HandlerFunc(handler.Authorize)},
	}
}
