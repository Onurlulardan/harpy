package billing

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func (handler *PlatformLicenseHTTPHandler) Routes() []httpedge.Route {
	if handler == nil {
		return nil
	}
	return []httpedge.Route{
		{
			Method: http.MethodGet, Pattern: "/api/platform/license",
			Handler: http.HandlerFunc(handler.readLicense),
		},
		{
			Method: http.MethodPost, Pattern: "/api/platform/license/refresh",
			Handler: http.HandlerFunc(handler.createRefresh),
		},
		{
			Method:  http.MethodGet,
			Pattern: "/api/platform/license/refresh-operations/{operation_id}",
			Handler: http.HandlerFunc(handler.readRefreshOperation),
		},
	}
}
