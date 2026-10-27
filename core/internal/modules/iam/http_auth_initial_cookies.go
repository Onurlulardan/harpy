package iam

import (
	"crypto/subtle"
	"net/http"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

const (
	initialPasswordCookieName = "__Secure-harpy-initial-change"
	initialPasswordCookiePath = "/api/auth/initial-password"
	initialPasswordCSRFName   = "__Host-harpy-initial-csrf"
)

func initialPasswordCookieValues(
	challenge AuthHTTPInitialChallenge,
	now time.Time,
) ([]string, error) {
	expiresAt := challenge.ExpiresAt().UTC().Truncate(time.Second)
	remaining := int(expiresAt.Sub(now).Seconds())
	if remaining < 1 || remaining > int(authentication.InitialPasswordLifetime.Seconds()) {
		return nil, authentication.ErrInitialPasswordState
	}
	var values []string
	err := challenge.RevealOnce(func(cookie, csrf []byte) error {
		if len(cookie) != 66 || len(csrf) != 43 {
			return authentication.ErrInitialPasswordState
		}
		values = []string{
			(&http.Cookie{Name: initialPasswordCookieName, Value: string(cookie),
				Path: initialPasswordCookiePath, Expires: expiresAt, MaxAge: remaining,
				HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode}).String(),
			(&http.Cookie{Name: initialPasswordCSRFName, Value: string(csrf),
				Path: "/", Expires: expiresAt, MaxAge: remaining,
				Secure: true, SameSite: http.SameSiteStrictMode}).String(),
		}
		return nil
	})
	if err != nil || len(values) != 2 {
		return nil, authentication.ErrInitialPasswordState
	}
	return values, nil
}

func presentedInitialPassword(
	r *http.Request,
) (*authentication.PresentedInitialPassword, bool) {
	cookie, cookieOK := requestCookie(r, initialPasswordCookieName, 66)
	csrfCookie, csrfCookieOK := requestCookie(r, initialPasswordCSRFName, 43)
	csrfHeader, csrfHeaderOK := exactCSRFHeader(r)
	defer clear(cookie)
	defer clear(csrfCookie)
	defer clear(csrfHeader)
	if !cookieOK || !csrfCookieOK || !csrfHeaderOK ||
		len(csrfCookie) != len(csrfHeader) ||
		subtle.ConstantTimeCompare(csrfCookie, csrfHeader) != 1 {
		return nil, false
	}
	presented, err := authentication.ParseInitialPassword(cookie, csrfCookie)
	return presented, err == nil
}

func expiredInitialPasswordCookieValues() []string {
	expired := time.Unix(1, 0).UTC()
	return []string{
		(&http.Cookie{Name: initialPasswordCookieName, Path: initialPasswordCookiePath,
			Expires: expired, MaxAge: -1, HttpOnly: true, Secure: true,
			SameSite: http.SameSiteStrictMode}).String(),
		(&http.Cookie{Name: initialPasswordCSRFName, Path: "/",
			Expires: expired, MaxAge: -1, Secure: true,
			SameSite: http.SameSiteStrictMode}).String(),
	}
}
