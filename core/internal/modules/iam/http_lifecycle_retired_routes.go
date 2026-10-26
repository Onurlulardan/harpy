package iam

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func (handler *LifecycleHTTP) retiredRoutes() []httpedge.Route {
	definitions := []struct{ method, pattern, replacement string }{
		{http.MethodPost, "/api/auth/register", "/api/invitations/accept"},
		{http.MethodPost, "/api/auth/forgot-password", "/api/auth/password/reset"},
		{http.MethodPost, "/api/auth/reset-password", "/api/auth/password/reset/complete"},
		{http.MethodPost, "/api/auth/portal/lookup", ""},
		{http.MethodPost, "/api/auth/portal/setup-password", "/api/auth/portal/setup/complete"},
		{http.MethodPost, "/api/auth/portal/forgot-password", "/api/auth/password/reset"},
		{http.MethodDelete, "/api/auth/portal-users/{id}/link", ""},
		{http.MethodPost, "/api/auth/validate", "/api/auth/session"},
		{http.MethodGet, "/api/auth/verify-email/{token}", "/api/auth/identifiers/verify"},
		{http.MethodPost, "/api/auth/create-verification-token", "/api/iam/identifiers/{id}/verification"},
		{http.MethodPost, "/api/auth/change-password", "/api/iam/password/change"},
		{http.MethodPost, "/api/profile/change-password", "/api/iam/password/change"},
		{http.MethodGet, "/api/auth/sessions", "/api/iam/sessions"},
		{http.MethodDelete, "/api/auth/sessions/{id}", "/api/iam/sessions/{id}"},
		{http.MethodDelete, "/api/auth/sessions", "/api/iam/sessions"},
		{http.MethodPost, "/api/auth/sessions/terminate-all", "/api/iam/sessions"},
		{http.MethodGet, "/api/auth/login-history", ""},
		{http.MethodGet, "/api/auth/admin/sessions", ""},
		{http.MethodDelete, "/api/auth/admin/sessions/{id}", ""},
		{http.MethodDelete, "/api/auth/admin/sessions", ""},
		{http.MethodPost, "/api/auth/admin/sessions/terminate-all", ""},
	}
	routes := make([]httpedge.Route, len(definitions))
	for index, definition := range definitions {
		replacement := definition.replacement
		routes[index] = httpedge.Route{Method: definition.method, Pattern: definition.pattern,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handler.retired(w, r, replacement)
			})}
	}
	return routes
}
