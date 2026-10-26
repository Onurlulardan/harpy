package publicationhttp

import (
	"net/http"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

const (
	publicationsPath    = "/api/analytics/publications"
	publicationPath     = publicationsPath + "/{release_id}"
	publicQueryPath     = "/api/published/queries/{public_id}/{slug}"
	publicDashboardPath = "/api/published/dashboards/{public_id}/{slug}"
)

func (handler *Handler) Routes() []httptransport.Route {
	return []httptransport.Route{
		{Method: http.MethodPost, Pattern: publicationsPath, Handler: http.HandlerFunc(handler.Create)},
		{Method: http.MethodGet, Pattern: publicationsPath, Handler: http.HandlerFunc(handler.List)},
		{Method: http.MethodGet, Pattern: publicationPath, Handler: http.HandlerFunc(handler.Get)},
		{Method: http.MethodPost, Pattern: publicationPath + "/submissions", Handler: http.HandlerFunc(handler.Submit)},
		{Method: http.MethodPost, Pattern: publicationPath + "/approvals", Handler: http.HandlerFunc(handler.Approve)},
		{Method: http.MethodPost, Pattern: publicationPath + "/activations", Handler: http.HandlerFunc(handler.Activate)},
		{Method: http.MethodPost, Pattern: publicationPath + "/revocations", Handler: http.HandlerFunc(handler.Revoke)},
		{Method: http.MethodGet, Pattern: publicQueryPath, Handler: http.HandlerFunc(handler.Execute)},
		{Method: http.MethodPost, Pattern: publicQueryPath, Handler: http.HandlerFunc(handler.Execute)},
		{Method: http.MethodGet, Pattern: publicDashboardPath, Handler: http.HandlerFunc(handler.Execute)},
	}
}
