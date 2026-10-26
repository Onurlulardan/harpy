package iam

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/csrf"
)

func TestAuthHTTPLogoutRevokesFamilyAndExpiresCookies(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	_, refreshCookie, csrfCookie := fixture.successfulLogin(t, false)
	request := fixture.browserRequest(
		http.MethodPost, "/api/auth/logout", refreshCookie, csrfCookie,
	)
	response := httptest.NewRecorder()
	fixture.handler.logout(response, request)
	if response.Code != http.StatusNoContent || response.Body.Len() != 0 ||
		fixture.app.logoutCalls != 1 || response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("logout status/calls=%d/%d body=%s", response.Code, fixture.app.logoutCalls, response.Body.String())
	}
	assertExpiredAuthCookies(t, response)
}

func TestAuthHTTPLogoutRejectsCSRFBeforeRevocation(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	_, refreshCookie, csrfCookie := fixture.successfulLogin(t, false)
	request := fixture.browserRequest(
		http.MethodPost, "/api/auth/logout", refreshCookie, csrfCookie,
	)
	request.Header.Add(csrf.HeaderName, csrfCookie.Value)
	response := httptest.NewRecorder()
	fixture.handler.logout(response, request)
	if response.Code != http.StatusForbidden || fixture.app.logoutCalls != 0 {
		t.Fatalf("logout CSRF status/calls=%d/%d", response.Code, fixture.app.logoutCalls)
	}
}

func TestAuthHTTPLogoutInvalidCredentialIsGenericAndCleared(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	_, refreshCookie, csrfCookie := fixture.successfulLogin(t, false)
	fixture.app.logoutError = ErrAuthHTTPInvalidCredentials
	request := fixture.browserRequest(
		http.MethodPost, "/api/auth/logout", refreshCookie, csrfCookie,
	)
	response := httptest.NewRecorder()
	fixture.handler.logout(response, request)
	if response.Code != http.StatusUnauthorized || fixture.app.logoutCalls != 1 {
		t.Fatalf("logout rejection status/calls=%d/%d", response.Code, fixture.app.logoutCalls)
	}
	assertExpiredAuthCookies(t, response)
}
