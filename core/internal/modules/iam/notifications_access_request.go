package iam

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/csrf"
)

func notificationsOrigin(runtime *NotificationsAccessRuntime,
	values []string, required bool,
) (string, bool) {
	if len(values) == 0 {
		return "", !required
	}
	origin, valid := exactNonempty(values, 2048)
	return origin, valid && runtime.shared.security.browserOriginAllowed(origin)
}

func notificationsCSRF(values, cookies []string) ([]byte, []byte, bool) {
	header, valid := exactNonempty(values, csrf.MaximumWireSize)
	if !valid {
		return nil, nil, false
	}
	request := &http.Request{Header: make(http.Header)}
	request.Header["Cookie"] = append([]string(nil), cookies...)
	cookie, valid := requestCookie(request, csrf.CSRFCookieName, csrf.MaximumWireSize)
	if !valid {
		return nil, nil, false
	}
	return cookie, []byte(header), true
}
