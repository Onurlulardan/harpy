package iam

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/csrf"
)

func TestAuthHTTPRefreshRotatesAccessRefreshAndCSRF(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	_, oldRefresh, oldCSRF := fixture.successfulLogin(t, false)
	fixture.app.refreshGrant = fixture.newGrant(t, fixture.claims)
	request := fixture.browserRequest(
		http.MethodPost, "/api/auth/refresh", oldRefresh, oldCSRF,
	)
	response := httptest.NewRecorder()
	fixture.handler.refresh(response, request)
	if response.Code != http.StatusOK || fixture.app.refreshCalls != 1 {
		t.Fatalf("refresh status/calls=%d/%d body=%s", response.Code, fixture.app.refreshCalls, response.Body.String())
	}
	document := decodeAuthResponse(t, response)
	if document["token"] == "" || document["expires_at"] == "" || len(document) != 2 {
		t.Fatalf("unexpected refresh response: %v", document)
	}
	newRefresh, newCSRF := responseAuthCookies(response)
	if newRefresh == nil || newCSRF == nil ||
		newRefresh.Value == oldRefresh.Value || newCSRF.Value == oldCSRF.Value {
		t.Fatalf("credentials were not rotated: %v", response.Header().Values("Set-Cookie"))
	}
	assertAuthCookieContract(t, response)
}

func TestAuthHTTPRefreshReplayCommitsThenClearsBrowserCredentials(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	_, oldRefresh, oldCSRF := fixture.successfulLogin(t, false)
	fixture.app.refreshError = ErrAuthHTTPInvalidCredentials
	request := fixture.browserRequest(
		http.MethodPost, "/api/auth/refresh", oldRefresh, oldCSRF,
	)
	response := httptest.NewRecorder()
	fixture.handler.refresh(response, request)
	if response.Code != http.StatusUnauthorized || fixture.app.refreshCalls != 1 {
		t.Fatalf("replay status/calls=%d/%d", response.Code, fixture.app.refreshCalls)
	}
	assertExpiredAuthCookies(t, response)
	if strings.Contains(response.Body.String(), "reuse") ||
		strings.Contains(response.Body.String(), "family") {
		t.Fatalf("replay detail leaked: %s", response.Body.String())
	}
}

func TestAuthHTTPRefreshRejectsCSRFBeforeApplication(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	_, refreshCookie, csrfCookie := fixture.successfulLogin(t, false)
	request := fixture.browserRequest(
		http.MethodPost, "/api/auth/refresh", refreshCookie, csrfCookie,
	)
	request.Header.Set(csrf.HeaderName, "different-proof")
	response := httptest.NewRecorder()
	fixture.handler.refresh(response, request)
	if response.Code != http.StatusForbidden || fixture.app.refreshCalls != 0 {
		t.Fatalf("CSRF status/calls=%d/%d", response.Code, fixture.app.refreshCalls)
	}
	document := decodeAuthResponse(t, response)
	if document["error"].(map[string]any)["code"] != "AUTH_CSRF_INVALID" {
		t.Fatalf("unexpected CSRF failure: %v", document)
	}
}

func TestAuthHTTPRefreshAcceptsOnlyAbsentOrEmptyObjectBody(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	_, refreshCookie, csrfCookie := fixture.successfulLogin(t, false)
	fixture.app.refreshGrant = fixture.newGrant(t, fixture.claims)
	request := fixture.browserRequest(
		http.MethodPost, "/api/auth/refresh", refreshCookie, csrfCookie,
	)
	request.Body = http.NoBody
	request.ContentLength = 0
	request.Header.Del("Content-Type")
	response := httptest.NewRecorder()
	fixture.handler.refresh(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("bodyless refresh status=%d", response.Code)
	}
	request = httptest.NewRequest(
		http.MethodPost, "/api/auth/refresh", strings.NewReader(`{"token":"forbidden"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	fixture.handler.refresh(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("non-empty refresh body status=%d", response.Code)
	}
}

func responseAuthCookies(
	response *httptest.ResponseRecorder,
) (*http.Cookie, *http.Cookie) {
	var refreshCookie, csrfCookie *http.Cookie
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == csrf.RefreshCookieName {
			refreshCookie = cookie
		} else if cookie.Name == csrf.CSRFCookieName {
			csrfCookie = cookie
		}
	}
	return refreshCookie, csrfCookie
}

func assertExpiredAuthCookies(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	refreshCookie, csrfCookie := responseAuthCookies(response)
	if refreshCookie == nil || csrfCookie == nil ||
		refreshCookie.MaxAge != -1 || csrfCookie.MaxAge != -1 {
		t.Fatalf("auth cookies not expired: %v", response.Header().Values("Set-Cookie"))
	}
}
