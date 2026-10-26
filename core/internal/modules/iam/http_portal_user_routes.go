package iam

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func (handler *PortalUserHTTP) Routes() []httpedge.Route {
	if handler == nil {
		return nil
	}
	return []httpedge.Route{
		{Method: http.MethodPost, Pattern: "/api/auth/portal-users", Handler: http.HandlerFunc(handler.create)},
		{Method: http.MethodPost, Pattern: "/api/auth/portal-users/{id}/invite", Handler: http.HandlerFunc(handler.invite)},
		{Method: http.MethodPatch, Pattern: "/api/auth/portal-users/{id}/status", Handler: http.HandlerFunc(handler.changeStatus)},
		{Method: http.MethodPost, Pattern: "/api/auth/portal-users/{id}/console-access", Handler: http.HandlerFunc(handler.grantConsole)},
		{Method: http.MethodDelete, Pattern: "/api/auth/portal-users/{id}/console-access", Handler: http.HandlerFunc(handler.revokeConsole)},
	}
}
