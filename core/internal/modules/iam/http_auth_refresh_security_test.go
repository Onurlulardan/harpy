package iam

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthHTTPRefreshBindsCSRFToExactRefreshSelector(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	_, firstRefresh, _ := fixture.successfulLogin(t, false)
	_, _, secondCSRF := fixture.successfulLogin(t, false)
	request := fixture.browserRequest(
		http.MethodPost, "/api/auth/refresh", firstRefresh, secondCSRF,
	)
	response := httptest.NewRecorder()
	fixture.handler.refresh(response, request)
	if response.Code != http.StatusForbidden || fixture.app.refreshCalls != 0 {
		t.Fatalf("selector substitution status/calls=%d/%d", response.Code, fixture.app.refreshCalls)
	}
}

func TestAuthHTTPRefreshRejectsDuplicateCredentialCookies(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	_, refreshCookie, csrfCookie := fixture.successfulLogin(t, false)
	request := fixture.browserRequest(
		http.MethodPost, "/api/auth/refresh", refreshCookie, csrfCookie,
	)
	request.AddCookie(refreshCookie)
	response := httptest.NewRecorder()
	fixture.handler.refresh(response, request)
	if response.Code != http.StatusUnauthorized || fixture.app.refreshCalls != 0 {
		t.Fatalf("duplicate cookie status/calls=%d/%d", response.Code, fixture.app.refreshCalls)
	}
	assertExpiredAuthCookies(t, response)
}
