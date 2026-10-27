package dashboardexporthttp

import (
	"net/http"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

const (
	exportsPath = "/api/analytics/dashboards/{dashboard_id}/exports"
	exportPath  = exportsPath + "/{export_id}"
)

func (handler *Handler) Routes() []httptransport.Route {
	return []httptransport.Route{
		{Method: http.MethodPost, Pattern: exportsPath, Handler: http.HandlerFunc(handler.Submit)},
		{Method: http.MethodGet, Pattern: exportPath, Handler: http.HandlerFunc(handler.Get)},
		{Method: http.MethodDelete, Pattern: exportPath, Handler: http.HandlerFunc(handler.Delete)},
		{Method: http.MethodGet, Pattern: exportPath + "/content", Handler: http.HandlerFunc(handler.Content)},
	}
}
