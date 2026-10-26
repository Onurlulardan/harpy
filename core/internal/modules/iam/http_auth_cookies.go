package iam

import (
	"net/http"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/csrf"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/refreshtoken"
)

func requestCookie(r *http.Request, name string, maximum int) ([]byte, bool) {
	if r == nil || name == "" {
		return nil, false
	}
	var value string
	count := 0
	for _, cookie := range r.Cookies() {
		if cookie.Name == name {
			value, count = cookie.Value, count+1
		}
	}
	if count != 1 || value == "" || len(value) > maximum {
		return nil, false
	}
	return []byte(value), true
}

func presentedRefresh(
	r *http.Request,
) (*refreshtoken.Presented, [16]byte, bool) {
	wire, valid := requestCookie(r, csrf.RefreshCookieName, 66)
	if !valid {
		return nil, [16]byte{}, false
	}
	presented, err := refreshtoken.Parse(wire)
	clear(wire)
	if err != nil {
		return nil, [16]byte{}, false
	}
	selector, err := presented.Selector()
	if err != nil {
		presented.Destroy()
		return nil, [16]byte{}, false
	}
	return presented, selector, true
}

func refreshCookieValue(
	wire []byte,
	absoluteExpiry time.Time,
	now time.Time,
) (string, bool) {
	expiresAt := absoluteExpiry.UTC().Truncate(time.Second)
	remaining := int(expiresAt.Sub(now).Seconds())
	if len(wire) != 66 || remaining < 1 {
		return "", false
	}
	cookie := http.Cookie{
		Name: csrf.RefreshCookieName, Value: string(wire),
		Path: csrf.RefreshCookiePath, Expires: expiresAt,
		MaxAge: remaining, HttpOnly: true, Secure: true,
		SameSite: http.SameSiteLaxMode,
	}
	value := cookie.String()
	return value, value != ""
}

func csrfCookieValue(wire []byte, now time.Time) (string, bool) {
	if len(wire) == 0 || len(wire) > csrf.MaximumWireSize || now.Nanosecond() != 0 {
		return "", false
	}
	cookie := http.Cookie{
		Name: csrf.CSRFCookieName, Value: string(wire), Path: csrf.CSRFCookiePath,
		Expires: now.Add(csrf.Lifetime), MaxAge: int(csrf.Lifetime.Seconds()),
		Secure: true, SameSite: http.SameSiteLaxMode,
	}
	value := cookie.String()
	return value, value != ""
}

func expiredAuthCookieValues() []string {
	expired := time.Unix(1, 0).UTC()
	return []string{
		(&http.Cookie{
			Name: csrf.RefreshCookieName, Path: csrf.RefreshCookiePath,
			Expires: expired, MaxAge: -1, HttpOnly: true, Secure: true,
			SameSite: http.SameSiteLaxMode,
		}).String(),
		(&http.Cookie{
			Name: csrf.CSRFCookieName, Path: csrf.CSRFCookiePath,
			Expires: expired, MaxAge: -1, Secure: true,
			SameSite: http.SameSiteLaxMode,
		}).String(),
	}
}

func appendSetCookies(header http.Header, values []string) {
	for _, value := range values {
		header.Add("Set-Cookie", value)
	}
}
