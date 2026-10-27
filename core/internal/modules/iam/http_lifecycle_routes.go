package iam

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func (handler *LifecycleHTTP) Routes() []httpedge.Route {
	if handler == nil {
		return nil
	}
	routes := []httpedge.Route{
		{Method: http.MethodPost, Pattern: "/api/iam/identifiers", Handler: http.HandlerFunc(handler.addIdentifier)},
		{Method: http.MethodPost, Pattern: "/api/iam/identifiers/{id}/verification", Handler: http.HandlerFunc(handler.startEmailVerification)},
		{Method: http.MethodPost, Pattern: "/api/auth/identifiers/verify", Handler: http.HandlerFunc(handler.verifyIdentifier)},
		{Method: http.MethodPost, Pattern: "/api/iam/identifiers/{id}/tombstone", Handler: http.HandlerFunc(handler.tombstoneIdentifier)},
		{Method: http.MethodPost, Pattern: "/api/iam/identifiers/{id}/retire", Handler: http.HandlerFunc(handler.retireIdentifier)},
		{Method: http.MethodPost, Pattern: "/api/auth/password/reset", Handler: http.HandlerFunc(handler.startPasswordReset)},
		{Method: http.MethodPost, Pattern: "/api/auth/password/reset/complete", Handler: http.HandlerFunc(handler.completePasswordReset)},
		{Method: http.MethodPost, Pattern: "/api/auth/initial-admin/setup/complete", Handler: http.HandlerFunc(handler.completeInitialAdminSetup)},
		{Method: http.MethodPost, Pattern: "/api/iam/password/change", Handler: http.HandlerFunc(handler.changePassword)},
		{Method: http.MethodGet, Pattern: "/api/iam/portal-setups/{binding_id}", Handler: http.HandlerFunc(handler.readPortalSetup)},
		{Method: http.MethodPost, Pattern: "/api/iam/portal-setups/{binding_id}/token", Handler: http.HandlerFunc(handler.startPortalSetup)},
		{Method: http.MethodPost, Pattern: "/api/auth/portal/setup/complete", Handler: http.HandlerFunc(handler.completePortalSetup)},
		{Method: http.MethodPost, Pattern: "/api/iam/portal-setups/{binding_id}/decision", Handler: http.HandlerFunc(handler.decidePortalSetup)},
		{Method: http.MethodPost, Pattern: "/api/iam/authenticators/webauthn/enrollments", Handler: http.HandlerFunc(handler.startWebAuthn)},
		{Method: http.MethodPost, Pattern: "/api/iam/authenticators/webauthn/enrollments/{authenticator_id}/complete", Handler: http.HandlerFunc(handler.completeWebAuthn)},
		{Method: http.MethodPost, Pattern: "/api/iam/authenticators/totp/enrollments", Handler: http.HandlerFunc(handler.startTOTP)},
		{Method: http.MethodPost, Pattern: "/api/iam/authenticators/totp/enrollments/{authenticator_id}/complete", Handler: http.HandlerFunc(handler.completeTOTP)},
		{Method: http.MethodPost, Pattern: "/api/iam/authenticators/recovery-codes", Handler: http.HandlerFunc(handler.issueRecoveryCodes)},
		{Method: http.MethodGet, Pattern: "/api/iam/authenticators/recovery-codes", Handler: http.HandlerFunc(handler.listRecoveryCodes)},
		{Method: http.MethodGet, Pattern: "/api/iam/authenticators", Handler: http.HandlerFunc(handler.listAuthenticators)},
		{Method: http.MethodPost, Pattern: "/api/iam/authenticators/{id}/suspend", Handler: http.HandlerFunc(handler.suspendAuthenticator)},
		{Method: http.MethodPost, Pattern: "/api/iam/authenticators/{id}/retire", Handler: http.HandlerFunc(handler.retireAuthenticator)},
		{Method: http.MethodGet, Pattern: "/api/iam/sessions", Handler: http.HandlerFunc(handler.listSessions)},
		{Method: http.MethodDelete, Pattern: "/api/iam/sessions/{id}", Handler: http.HandlerFunc(handler.revokeSession)},
		{Method: http.MethodDelete, Pattern: "/api/iam/sessions", Handler: http.HandlerFunc(handler.revokeAllSessions)},
	}
	return append(routes, handler.retiredRoutes()...)
}
