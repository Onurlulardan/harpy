package iam

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func (handler *IdentityHTTP) Routes() []httpedge.Route {
	if handler == nil {
		return nil
	}
	return []httpedge.Route{
		{Method: http.MethodGet, Pattern: "/api/users", Handler: http.HandlerFunc(handler.listUsers)},
		{Method: http.MethodPost, Pattern: "/api/users", Handler: http.HandlerFunc(handler.createUser)},
		{Method: http.MethodDelete, Pattern: "/api/users", Handler: http.HandlerFunc(handler.deleteUsers)},
		{Method: http.MethodPost, Pattern: "/api/users/query", Handler: http.HandlerFunc(handler.queryUsers)},
		{Method: http.MethodGet, Pattern: "/api/users/{id}", Handler: http.HandlerFunc(handler.getUser)},
		{Method: http.MethodPut, Pattern: "/api/users/{id}", Handler: http.HandlerFunc(handler.updateUser)},
		{Method: http.MethodDelete, Pattern: "/api/users/{id}", Handler: http.HandlerFunc(handler.deleteUser)},
		{Method: http.MethodPatch, Pattern: "/api/users/{id}/state", Handler: http.HandlerFunc(handler.changeUserState)},
		{Method: http.MethodGet, Pattern: "/api/roles", Handler: http.HandlerFunc(handler.listRoles)},
		{Method: http.MethodPost, Pattern: "/api/roles", Handler: http.HandlerFunc(handler.createRole)},
		{Method: http.MethodDelete, Pattern: "/api/roles", Handler: http.HandlerFunc(handler.deleteRoles)},
		{Method: http.MethodPost, Pattern: "/api/roles/query", Handler: http.HandlerFunc(handler.queryRoles)},
		{Method: http.MethodGet, Pattern: "/api/roles/{id}", Handler: http.HandlerFunc(handler.getRole)},
		{Method: http.MethodPut, Pattern: "/api/roles/{id}", Handler: http.HandlerFunc(handler.updateRole)},
		{Method: http.MethodDelete, Pattern: "/api/roles/{id}", Handler: http.HandlerFunc(handler.deleteRole)},
		{Method: http.MethodPatch, Pattern: "/api/roles/{id}/state", Handler: http.HandlerFunc(handler.changeRoleState)},
		{Method: http.MethodGet, Pattern: "/api/organizations", Handler: http.HandlerFunc(handler.listOrganizations)},
		{Method: http.MethodPost, Pattern: "/api/organizations", Handler: http.HandlerFunc(handler.createOrganization)},
		{Method: http.MethodDelete, Pattern: "/api/organizations", Handler: http.HandlerFunc(handler.deleteOrganizations)},
		{Method: http.MethodPost, Pattern: "/api/organizations/query", Handler: http.HandlerFunc(handler.queryOrganizations)},
		{Method: http.MethodGet, Pattern: "/api/organizations/{id}", Handler: http.HandlerFunc(handler.getOrganization)},
		{Method: http.MethodPut, Pattern: "/api/organizations/{id}", Handler: http.HandlerFunc(handler.updateOrganization)},
		{Method: http.MethodDelete, Pattern: "/api/organizations/{id}", Handler: http.HandlerFunc(handler.deleteOrganization)},
		{Method: http.MethodPatch, Pattern: "/api/organizations/{id}/state", Handler: http.HandlerFunc(handler.changeOrganizationState)},
	}
}
