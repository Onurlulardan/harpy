package iam

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

func (handler *AuthHTTP) executeLogin(
	w http.ResponseWriter,
	r *http.Request,
	input AuthHTTPLoginInput,
) {
	defer clear(input.Password)
	origin, valid := exactOrigin(r)
	if !valid || !handler.security.browserOriginAllowed(origin) {
		handler.writeAuthFailure(w, r, authFailureCSRF)
		return
	}
	_, valid = handler.authHTTPNow()
	if !valid {
		handler.report(errors.New("IAM auth HTTP clock is invalid"))
		handler.writeAuthFailure(w, r, authFailureUnavailable)
		return
	}
	if handler.executeInitialPasswordStart(w, r, input, origin) {
		return
	}
	grant, err := handler.application.Login(
		withAuthHTTPPeer(r.Context(), r.RemoteAddr), input,
	)
	if err != nil {
		if !nilAuthHTTPGrant(grant) {
			grant.Destroy()
		}
		handler.writeApplicationFailure(w, r, err)
		return
	}
	if nilAuthHTTPGrant(grant) {
		handler.report(errors.New("IAM login returned no grant"))
		handler.writeAuthFailure(w, r, authFailureUnavailable)
		return
	}
	defer grant.Destroy()
	now, valid := handler.authHTTPNow()
	if !valid {
		handler.report(errors.New("IAM login presentation clock is invalid"))
		handler.writeAuthFailure(w, r, authFailureUnavailable)
		return
	}
	response, cookies, err := handler.prepareGrant(
		r.Context(), grant, origin, now,
	)
	if err != nil {
		handler.report(errors.New("IAM login grant could not be presented"))
		handler.writeAuthFailure(w, r, authFailureUnavailable)
		return
	}
	appendSetCookies(w.Header(), cookies)
	if err := httpwire.WriteJSON(
		w, http.StatusOK, authHTTPCorrelationID(r), response, false,
	); err != nil {
		handler.report(fmt.Errorf("write IAM login response: %w", err))
	}
}

func (handler *AuthHTTP) authHTTPNow() (time.Time, bool) {
	now := handler.clock().UTC().Truncate(time.Second)
	return now, !now.IsZero() && now.Unix() > 0
}
