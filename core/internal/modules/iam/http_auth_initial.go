package iam

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

func (handler *AuthHTTP) initialPassword(w http.ResponseWriter, r *http.Request) {
	input, err := decodeInitialPassword(w, r)
	if err != nil {
		handler.writeInputFailure(w, r, err)
		return
	}
	defer clear(input.NewPassword)
	origin, valid := exactOrigin(r)
	if !valid || !handler.security.browserOriginAllowed(origin) {
		handler.writeAuthFailure(w, r, authFailureCSRF)
		return
	}
	presented, valid := presentedInitialPassword(r)
	if !valid {
		handler.writeAuthFailure(w, r, authFailureCSRF)
		return
	}
	defer presented.Destroy()
	grant, err := handler.application.CompleteInitialPassword(
		withAuthHTTPPeer(r.Context(), r.RemoteAddr), input, presented,
	)
	if err != nil {
		if !nilAuthHTTPGrant(grant) {
			grant.Destroy()
		}
		handler.writeApplicationFailure(w, r, err)
		return
	}
	if nilAuthHTTPGrant(grant) {
		handler.report(errors.New("IAM initial password returned no grant"))
		handler.writeAuthFailure(w, r, authFailureUnavailable)
		return
	}
	defer grant.Destroy()
	now, valid := handler.authHTTPNow()
	if !valid {
		handler.writeAuthFailure(w, r, authFailureUnavailable)
		return
	}
	response, cookies, err := handler.prepareGrant(r.Context(), grant, origin, now)
	if err != nil {
		handler.report(errors.New("IAM initial password grant presentation failed"))
		handler.writeAuthFailure(w, r, authFailureUnavailable)
		return
	}
	appendSetCookies(w.Header(), cookies)
	appendSetCookies(w.Header(), expiredInitialPasswordCookieValues())
	if err := httpwire.WriteJSON(
		w, http.StatusOK, authHTTPCorrelationID(r), response, false,
	); err != nil {
		handler.report(fmt.Errorf("write IAM initial password grant: %w", err))
	}
}
