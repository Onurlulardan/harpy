package iam

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

func (handler *AuthHTTP) currentSession(w http.ResponseWriter, r *http.Request) {
	now, valid := handler.authHTTPNow()
	if !valid {
		handler.report(errors.New("IAM current-session clock is invalid"))
		handler.writeAuthFailure(w, r, authFailureUnavailable)
		return
	}
	token, valid := exactBearer(r)
	if !valid {
		handler.writeAuthFailure(w, r, authFailureAuthentication)
		return
	}
	claims, err := handler.security.accessVerifier.Verify(r.Context(), token, now)
	if err != nil || !optionalTenantIDMatches(
		r.Header.Values(authTenantIDHeader), claims.Tenant(),
	) {
		handler.writeAuthFailure(w, r, authFailureAuthentication)
		return
	}
	view, err := handler.application.CurrentSession(r.Context(), claims)
	if err != nil {
		handler.writeApplicationFailure(w, r, err)
		return
	}
	response, valid := currentSessionResponse(view, claims)
	if !valid || !tenantHeadersMatch(r, view.TenantID, view.TenantSlug) {
		handler.writeAuthFailure(w, r, authFailureAuthentication)
		return
	}
	if err := httpwire.WriteJSON(
		w, http.StatusOK, authHTTPCorrelationID(r), response, false,
	); err != nil {
		handler.report(fmt.Errorf("write IAM current-session response: %w", err))
	}
}
