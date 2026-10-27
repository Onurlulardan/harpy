package iam

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func (handler *PermissionRequestHTTP) Routes() []httpedge.Route {
	if handler == nil {
		return nil
	}
	return []httpedge.Route{
		{Method: http.MethodGet, Pattern: "/api/permission-requests", Handler: http.HandlerFunc(handler.list)},
		{Method: http.MethodPost, Pattern: "/api/permission-requests", Handler: http.HandlerFunc(handler.create)},
		{Method: http.MethodPost, Pattern: "/api/permission-requests/bulk", Handler: http.HandlerFunc(handler.bulkCreate)},
		{Method: http.MethodGet, Pattern: "/api/permission-requests/mine", Handler: http.HandlerFunc(handler.list)},
		{Method: http.MethodGet, Pattern: "/api/permission-requests/catalog/resources", Handler: http.HandlerFunc(handler.listCatalogResources)},
		{Method: http.MethodGet, Pattern: "/api/permission-requests/catalog/actions", Handler: http.HandlerFunc(handler.listCatalogActions)},
		{Method: http.MethodGet, Pattern: "/api/permission-requests/catalog/missing-resources", Handler: http.HandlerFunc(handler.listMissingCatalogResources)},
		{Method: http.MethodGet, Pattern: "/api/permission-requests/{id}", Handler: http.HandlerFunc(handler.get)},
		{Method: http.MethodPost, Pattern: "/api/permission-requests/{id}/review", Handler: http.HandlerFunc(handler.review)},
		{Method: http.MethodPost, Pattern: "/api/permission-requests/{id}/approve", Handler: http.HandlerFunc(handler.approve)},
		{Method: http.MethodPost, Pattern: "/api/permission-requests/{id}/reject", Handler: http.HandlerFunc(handler.reject)},
		{Method: http.MethodPost, Pattern: "/api/permission-requests/{id}/cancel", Handler: http.HandlerFunc(handler.cancel)},
		{Method: http.MethodPost, Pattern: "/api/permission-requests/{id}/expire", Handler: http.HandlerFunc(handler.expire)},
	}
}
