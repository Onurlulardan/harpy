package iam

import (
	"errors"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/csrf"
)

func (handler *AuthHTTP) csrf(w http.ResponseWriter, r *http.Request) {
	origin, valid := exactOrigin(r)
	if !valid || !handler.security.browserOriginAllowed(origin) {
		handler.writeAuthFailure(w, r, authFailureCSRF)
		return
	}
	now, valid := handler.authHTTPNow()
	if !valid {
		handler.report(errors.New("IAM CSRF clock is invalid"))
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
	if err := handler.application.ValidateRefresh(
		withAuthHTTPPeer(r.Context(), r.RemoteAddr), presented,
	); err != nil {
		if authHTTPRejected(err) {
			appendSetCookies(w.Header(), expiredAuthCookieValues())
		}
		handler.writeApplicationFailure(w, r, err)
		return
	}
	now, valid = handler.authHTTPNow()
	if !valid {
		handler.report(errors.New("IAM CSRF presentation clock is invalid"))
		handler.writeAuthFailure(w, r, authFailureUnavailable)
		return
	}
	credential, err := handler.security.csrfSigner.Sign(selector, origin, now)
	if err != nil {
		handler.report(errors.New("IAM CSRF credential could not be signed"))
		handler.writeAuthFailure(w, r, authFailureUnavailable)
		return
	}
	defer credential.Destroy()
	var cookie string
	err = credential.RevealOnce(func(wire []byte) error {
		value, ok := csrfCookieValue(wire, now)
		if !ok {
			return csrf.ErrInvalidCredential
		}
		cookie = value
		return nil
	})
	if err != nil || cookie == "" {
		handler.report(errors.New("IAM CSRF credential could not be presented"))
		handler.writeAuthFailure(w, r, authFailureUnavailable)
		return
	}
	appendSetCookies(w.Header(), []string{cookie})
	writeAuthNoContent(w, r)
}
