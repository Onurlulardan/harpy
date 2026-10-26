package iam

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

func (handler *AuthHTTP) refresh(w http.ResponseWriter, r *http.Request) {
	if err := decodeOptionalEmptyObject(w, r); err != nil {
		handler.writeInputFailure(w, r, err)
		return
	}
	now, valid := handler.authHTTPNow()
	if !valid {
		handler.report(errors.New("IAM refresh clock is invalid"))
		handler.writeAuthFailure(w, r, authFailureUnavailable)
		return
	}
	presented, selector, valid := presentedRefresh(r)
	if !valid {
		appendSetCookies(w.Header(), expiredAuthCookieValues())
		handler.writeAuthFailure(w, r, authFailureAuthentication)
		return
	}
	defer presented.Destroy()
	origin, valid := handler.verifyCSRFProof(r, selector, now)
	if !valid {
		handler.writeAuthFailure(w, r, authFailureCSRF)
		return
	}
	grant, err := handler.application.Refresh(
		withAuthHTTPPeer(r.Context(), r.RemoteAddr), presented,
	)
	if err != nil {
		if !nilAuthHTTPGrant(grant) {
			grant.Destroy()
		}
		if authHTTPRejected(err) {
			appendSetCookies(w.Header(), expiredAuthCookieValues())
		}
		handler.writeApplicationFailure(w, r, err)
		return
	}
	if nilAuthHTTPGrant(grant) {
		handler.report(errors.New("IAM refresh returned no grant"))
		handler.writeAuthFailure(w, r, authFailureUnavailable)
		return
	}
	defer grant.Destroy()
	now, valid = handler.authHTTPNow()
	if !valid {
		handler.report(errors.New("IAM refresh presentation clock is invalid"))
		handler.writeAuthFailure(w, r, authFailureUnavailable)
		return
	}
	response, cookies, err := handler.prepareGrant(r.Context(), grant, origin, now)
	if err != nil {
		handler.report(errors.New("IAM refresh grant could not be presented"))
		handler.writeAuthFailure(w, r, authFailureUnavailable)
		return
	}
	appendSetCookies(w.Header(), cookies)
	if err := httpwire.WriteJSON(
		w, http.StatusOK, authHTTPCorrelationID(r), response, false,
	); err != nil {
		handler.report(fmt.Errorf("write IAM refresh response: %w", err))
	}
}
