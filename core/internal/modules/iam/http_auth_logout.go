package iam

import (
	"errors"
	"net/http"
)

func (handler *AuthHTTP) logout(w http.ResponseWriter, r *http.Request) {
	if err := decodeOptionalEmptyObject(w, r); err != nil {
		handler.writeInputFailure(w, r, err)
		return
	}
	now, valid := handler.authHTTPNow()
	if !valid {
		handler.report(errors.New("IAM logout clock is invalid"))
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
	if _, valid := handler.verifyCSRFProof(r, selector, now); !valid {
		handler.writeAuthFailure(w, r, authFailureCSRF)
		return
	}
	if err := handler.application.Logout(
		withAuthHTTPPeer(r.Context(), r.RemoteAddr), presented,
	); err != nil {
		if authHTTPRejected(err) {
			appendSetCookies(w.Header(), expiredAuthCookieValues())
		}
		handler.writeApplicationFailure(w, r, err)
		return
	}
	appendSetCookies(w.Header(), expiredAuthCookieValues())
	writeAuthNoContent(w, r)
}
