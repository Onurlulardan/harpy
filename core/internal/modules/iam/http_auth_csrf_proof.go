package iam

import (
	"net/http"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/csrf"
)

func (handler *AuthHTTP) verifyCSRFProof(
	r *http.Request,
	selector [16]byte,
	now time.Time,
) (string, bool) {
	origin, originValid := exactOrigin(r)
	cookie, cookieValid := requestCookie(r, csrf.CSRFCookieName, csrf.MaximumWireSize)
	header, headerValid := exactCSRFHeader(r)
	defer clear(cookie)
	defer clear(header)
	if !originValid || !cookieValid || !headerValid {
		return "", false
	}
	if err := handler.security.csrfVerifier.Verify(
		r.Context(), cookie, header, selector, origin, now,
	); err != nil {
		return "", false
	}
	return origin, true
}
