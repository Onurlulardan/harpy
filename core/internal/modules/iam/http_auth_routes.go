package iam

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func (handler *AuthHTTP) Routes() []httpedge.Route {
	if handler == nil {
		return nil
	}
	return []httpedge.Route{
		{Method: http.MethodPost, Pattern: "/api/auth/login", Handler: http.HandlerFunc(handler.login)},
		{Method: http.MethodPost, Pattern: "/api/auth/initial-password", Handler: http.HandlerFunc(handler.initialPassword)},
		{Method: http.MethodPost, Pattern: "/api/auth/portal/login", Handler: http.HandlerFunc(handler.portalLogin)},
		{Method: http.MethodPost, Pattern: "/api/auth/refresh", Handler: http.HandlerFunc(handler.refresh)},
		{Method: http.MethodPost, Pattern: "/api/auth/logout", Handler: http.HandlerFunc(handler.logout)},
		{Method: http.MethodGet, Pattern: "/api/auth/csrf", Handler: http.HandlerFunc(handler.csrf)},
		{Method: http.MethodPost, Pattern: "/api/auth/csrf", Handler: http.HandlerFunc(handler.csrf)},
		{Method: http.MethodGet, Pattern: "/api/auth/session", Handler: http.HandlerFunc(handler.currentSession)},
		{Method: http.MethodPost, Pattern: "/api/auth/reauthentication/start", Handler: http.HandlerFunc(handler.startReauthentication)},
		{Method: http.MethodPost, Pattern: "/api/auth/reauthentication/complete", Handler: http.HandlerFunc(handler.completeReauthentication)},
	}
}
