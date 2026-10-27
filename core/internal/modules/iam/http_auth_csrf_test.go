package iam

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/csrf"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func TestAuthHTTPCSFRReadOnlyValidationRenewsOnlyCSRFCookie(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	_, refreshCookie, oldCSRF := fixture.successfulLogin(t, false)
	request := fixture.browserRequest(
		http.MethodGet, "/api/auth/csrf", refreshCookie, nil,
	)
	response := httptest.NewRecorder()
	fixture.handler.csrf(response, request)
	if response.Code != http.StatusNoContent || fixture.app.validationCalls != 1 ||
		fixture.app.refreshCalls != 0 || fixture.app.logoutCalls != 0 {
		t.Fatalf("CSRF renewal status/validation=%d/%d", response.Code, fixture.app.validationCalls)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != csrf.CSRFCookieName ||
		cookies[0].Value == oldCSRF.Value || cookies[0].HttpOnly ||
		!cookies[0].Secure || cookies[0].Path != "/" {
		t.Fatalf("unexpected CSRF renewal cookie: %+v", cookies)
	}
}

func TestAuthHTTPCSFRBrowserPostRenewalUsesExactOrigin(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	_, refreshCookie, _ := fixture.successfulLogin(t, false)
	routes, err := httpedge.NewRoutes(func(error) {}, fixture.handler.Routes()...)
	if err != nil {
		t.Fatal(err)
	}
	request := fixture.browserRequest(
		http.MethodPost, "/api/auth/csrf", refreshCookie, nil,
	)
	response := httptest.NewRecorder()
	routes.ServeHTTP(response, request)
	cookies := response.Result().Cookies()
	if response.Code != http.StatusNoContent || fixture.app.validationCalls != 1 ||
		len(cookies) != 1 || cookies[0].Name != csrf.CSRFCookieName {
		t.Fatalf(
			"POST CSRF renewal status/validation/cookies=%d/%d/%+v",
			response.Code, fixture.app.validationCalls, cookies,
		)
	}
}

func TestAuthHTTPCSFRRejectsOriginBeforeCredentialRead(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	_, refreshCookie, _ := fixture.successfulLogin(t, false)
	request := fixture.browserRequest(
		http.MethodGet, "/api/auth/csrf", refreshCookie, nil,
	)
	request.Header["Origin"] = []string{"null"}
	response := httptest.NewRecorder()
	fixture.handler.csrf(response, request)
	if response.Code != http.StatusForbidden || fixture.app.validationCalls != 0 {
		t.Fatalf("origin status/validation=%d/%d", response.Code, fixture.app.validationCalls)
	}
}

func TestAuthHTTPCSFRInvalidRefreshIsGenericAndCleared(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	_, refreshCookie, _ := fixture.successfulLogin(t, false)
	fixture.app.validationError = ErrAuthHTTPInvalidCredentials
	request := fixture.browserRequest(
		http.MethodGet, "/api/auth/csrf", refreshCookie, nil,
	)
	response := httptest.NewRecorder()
	fixture.handler.csrf(response, request)
	if response.Code != http.StatusUnauthorized || fixture.app.validationCalls != 1 {
		t.Fatalf("validation status/calls=%d/%d", response.Code, fixture.app.validationCalls)
	}
	assertExpiredAuthCookies(t, response)
}
