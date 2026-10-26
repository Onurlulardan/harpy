package iam

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func (handler *RBACHTTP) Routes() []httpedge.Route {
	if handler == nil {
		return nil
	}
	return []httpedge.Route{
		{Method: http.MethodGet, Pattern: "/api/permissions", Handler: http.HandlerFunc(handler.listPermissions)},
		{Method: http.MethodPost, Pattern: "/api/permissions", Handler: http.HandlerFunc(handler.createPermission)},
		{Method: http.MethodGet, Pattern: "/api/permissions/resources", Handler: http.HandlerFunc(handler.listResources)},
		{Method: http.MethodPost, Pattern: "/api/permissions/resources", Handler: http.HandlerFunc(handler.createResource)},
		{Method: http.MethodGet, Pattern: "/api/permissions/resources/{id}", Handler: http.HandlerFunc(handler.getResource)},
		{Method: http.MethodPut, Pattern: "/api/permissions/resources/{id}", Handler: http.HandlerFunc(handler.updateResource)},
		{Method: http.MethodDelete, Pattern: "/api/permissions/resources/{id}", Handler: http.HandlerFunc(handler.deleteResource)},
		{Method: http.MethodGet, Pattern: "/api/permissions/actions", Handler: http.HandlerFunc(handler.listActions)},
		{Method: http.MethodPost, Pattern: "/api/permissions/actions", Handler: http.HandlerFunc(handler.createAction)},
		{Method: http.MethodGet, Pattern: "/api/permissions/actions/{id}", Handler: http.HandlerFunc(handler.getAction)},
		{Method: http.MethodPut, Pattern: "/api/permissions/actions/{id}", Handler: http.HandlerFunc(handler.updateAction)},
		{Method: http.MethodDelete, Pattern: "/api/permissions/actions/{id}", Handler: http.HandlerFunc(handler.deleteAction)},
		{Method: http.MethodPost, Pattern: "/api/permissions/check", Handler: http.HandlerFunc(handler.check)},
		{Method: http.MethodPost, Pattern: "/api/permissions/batch-check", Handler: http.HandlerFunc(handler.batchCheck)},
		{Method: http.MethodGet, Pattern: "/api/permissions/{id}", Handler: http.HandlerFunc(handler.getPermission)},
		{Method: http.MethodPut, Pattern: "/api/permissions/{id}", Handler: http.HandlerFunc(handler.updatePermission)},
		{Method: http.MethodDelete, Pattern: "/api/permissions/{id}", Handler: http.HandlerFunc(handler.deletePermission)},
		{Method: http.MethodGet, Pattern: "/api/users/{id}/permissions", Handler: http.HandlerFunc(handler.userEffective)},
		{Method: http.MethodGet, Pattern: "/api/roles/{id}/permissions", Handler: http.HandlerFunc(handler.roleEffective)},
		{Method: http.MethodGet, Pattern: "/api/organizations/{id}/permissions", Handler: http.HandlerFunc(handler.organizationEffective)},
		{Method: http.MethodGet, Pattern: "/api/permissions/cache/stats", Handler: http.HandlerFunc(handler.cacheRetired)},
		{Method: http.MethodPost, Pattern: "/api/permissions/cache/invalidate/{user_id}", Handler: http.HandlerFunc(handler.cacheRetired)},
		{Method: http.MethodPost, Pattern: "/api/permissions/cache/invalidate/role/{role_id}", Handler: http.HandlerFunc(handler.cacheRetired)},
		{Method: http.MethodPost, Pattern: "/api/permissions/cache/invalidate/org/{org_id}", Handler: http.HandlerFunc(handler.cacheRetired)},
		{Method: http.MethodPost, Pattern: "/api/permissions/cache/invalidate/tenant/{tenant_id}", Handler: http.HandlerFunc(handler.cacheRetired)},
		{Method: http.MethodPost, Pattern: "/api/permissions/cache/invalidate/all", Handler: http.HandlerFunc(handler.cacheRetired)},
	}
}
