package dashboardhttp

import (
	"net/http"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

const (
	dashboardsPath = "/api/analytics/dashboards"
	defaultsPath   = "/api/analytics/dashboard-defaults"
	dashboardPath  = dashboardsPath + "/{dashboard_id}"
)

func (handler *Handler) Routes() []httptransport.Route {
	return []httptransport.Route{
		{Method: http.MethodPost, Pattern: dashboardsPath, Handler: http.HandlerFunc(handler.Create)},
		{Method: http.MethodGet, Pattern: dashboardsPath, Handler: http.HandlerFunc(handler.List)},
		{Method: http.MethodGet, Pattern: defaultsPath + "/{surface}", Handler: http.HandlerFunc(handler.ReadDefault)},
		{Method: http.MethodGet, Pattern: dashboardPath, Handler: http.HandlerFunc(handler.Get)},
		{Method: http.MethodPatch, Pattern: dashboardPath, Handler: http.HandlerFunc(handler.Update)},
		{Method: http.MethodDelete, Pattern: dashboardPath, Handler: http.HandlerFunc(handler.Delete)},
		{Method: http.MethodPost, Pattern: dashboardPath + "/publications", Handler: http.HandlerFunc(handler.Publish)},
		{Method: http.MethodPost, Pattern: dashboardPath + "/acl-grants", Handler: http.HandlerFunc(handler.Grant)},
		{Method: http.MethodPost, Pattern: dashboardPath + "/acl-revocations", Handler: http.HandlerFunc(handler.Revoke)},
		{Method: http.MethodPut, Pattern: dashboardPath + "/defaults/{surface}", Handler: http.HandlerFunc(handler.Set)},
	}
}
