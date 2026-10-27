package queryhttp

import (
	"net/http"

	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

const (
	queriesPath = "/api/analytics/queries"
	queryPath   = queriesPath + "/{query_id}"
)

func (handler *Handler) Routes() []httptransport.Route {
	return []httptransport.Route{
		{Method: http.MethodPost, Pattern: queriesPath, Handler: http.HandlerFunc(handler.Create)},
		{Method: http.MethodGet, Pattern: queriesPath, Handler: http.HandlerFunc(handler.List)},
		{Method: http.MethodGet, Pattern: queryPath, Handler: http.HandlerFunc(handler.Get)},
		{Method: http.MethodGet, Pattern: queryPath + "/versions/latest-published",
			Handler: http.HandlerFunc(handler.ReadPublished)},
		{Method: http.MethodPatch, Pattern: queryPath, Handler: http.HandlerFunc(handler.Update)},
		{Method: http.MethodDelete, Pattern: queryPath, Handler: http.HandlerFunc(handler.Delete)},
		{Method: http.MethodPost, Pattern: queryPath + "/publications", Handler: http.HandlerFunc(handler.Publish)},
		{Method: http.MethodPost, Pattern: queryPath + "/revalidations", Handler: http.HandlerFunc(handler.Revalidate)},
		{Method: http.MethodPost, Pattern: queryPath + "/executions", Handler: http.HandlerFunc(handler.Execute)},
	}
}
