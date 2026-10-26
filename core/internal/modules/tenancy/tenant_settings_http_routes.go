package tenancy

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func (handler *TenantSettingsHTTP) Routes() []httpedge.Route {
	if !handler.Ready() {
		return nil
	}
	return []httpedge.Route{
		{Method: http.MethodGet, Pattern: "/api/tenant/settings",
			Handler: http.HandlerFunc(handler.getTenantSettings)},
		{Method: http.MethodPut, Pattern: "/api/tenant/settings",
			Handler: http.HandlerFunc(handler.updateTenantSettings)},
		{Method: http.MethodPost, Pattern: "/api/tenant/settings/test-email",
			Handler: http.HandlerFunc(handler.testTenantEmail)},
	}
}
