package iam

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
)

func (handler *AuthHTTP) reauthenticationClaims(
	w http.ResponseWriter,
	r *http.Request,
) (accesstoken.Claims, bool) {
	now, valid := handler.authHTTPNow()
	if !valid {
		handler.writeAuthFailure(w, r, authFailureUnavailable)
		return accesstoken.Claims{}, false
	}
	token, valid := exactBearer(r)
	if !valid {
		handler.writeAuthFailure(w, r, authFailureAuthentication)
		return accesstoken.Claims{}, false
	}
	claims, err := handler.security.accessVerifier.Verify(r.Context(), token, now)
	if err != nil || !optionalTenantIDMatches(
		r.Header.Values(authTenantIDHeader), claims.Tenant(),
	) {
		handler.writeAuthFailure(w, r, authFailureAuthentication)
		return accesstoken.Claims{}, false
	}
	view, err := handler.application.CurrentSession(r.Context(), claims)
	if err != nil || !tenantHeadersMatch(r, view.TenantID, view.TenantSlug) ||
		view.UserID != claims.Subject() || view.SessionID != claims.Session() {
		handler.writeApplicationFailure(w, r, errOrInvalidSession(err))
		return accesstoken.Claims{}, false
	}
	return claims, true
}

func errOrInvalidSession(err error) error {
	if err != nil {
		return err
	}
	return ErrAuthHTTPInvalidSession
}
