package api

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

const (
	layoutHTTPCollectionPath   = "/api/system-resources/{resource}/layout-revisions"
	layoutHTTPRevisionPath     = layoutHTTPCollectionPath + "/{revision_id}"
	layoutHTTPPublicationPath  = layoutHTTPRevisionPath + "/publication"
	layoutHTTPPresentationPath = "/api/system-resources/{resource}/presentation"
)

func (handler *systemResourceLayoutHTTPHandler) Routes() []httpedge.Route {
	if handler == nil {
		return nil
	}
	return []httpedge.Route{
		{Method: http.MethodGet, Pattern: layoutHTTPCollectionPath,
			Handler: http.HandlerFunc(handler.listRevisions)},
		{Method: http.MethodPut, Pattern: layoutHTTPRevisionPath,
			Handler: http.HandlerFunc(handler.putRevision)},
		{Method: http.MethodGet, Pattern: layoutHTTPRevisionPath,
			Handler: http.HandlerFunc(handler.getRevision)},
		{Method: http.MethodPatch, Pattern: layoutHTTPRevisionPath,
			Handler: http.HandlerFunc(handler.patchRevision)},
		{Method: http.MethodDelete, Pattern: layoutHTTPRevisionPath,
			Handler: http.HandlerFunc(handler.deleteRevision)},
		{Method: http.MethodPut, Pattern: layoutHTTPPublicationPath,
			Handler: http.HandlerFunc(handler.putPublication)},
		{Method: http.MethodGet, Pattern: layoutHTTPPresentationPath,
			Handler: http.HandlerFunc(handler.getPresentation)},
	}
}
