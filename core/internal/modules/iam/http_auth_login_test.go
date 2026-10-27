package iam

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/csrf"
)

func TestAuthHTTPLoginIssuesAccessAndExactBrowserCookies(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	grant := fixture.newGrant(t, fixture.claims)
	fixture.app.loginGrant = grant
	request := httptest.NewRequest(
		http.MethodPost, "/api/auth/login",
		strings.NewReader(`{"tenant_slug":"tenant-a","email":"Admin@Example.com","password":"valid-password"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", fixture.origin)
	response := httptest.NewRecorder()
	fixture.handler.login(response, request)
	if response.Code != http.StatusOK || fixture.app.loginCalls != 1 ||
		fixture.app.loginInput.Mode != AuthHTTPLoginOperator ||
		fixture.app.loginInput.TenantSlug != "tenant-a" ||
		fixture.app.loginInput.Identifier != "Admin@Example.com" {
		t.Fatalf("unexpected login result: status=%d input=%+v", response.Code, fixture.app.loginInput)
	}
	if !allCleared(fixture.app.loginInput.Password) || !grant.destroyed {
		t.Fatal("login material was not destroyed")
	}
	document := decodeAuthResponse(t, response)
	if document["token"] == "" || document["expires_at"] == "" ||
		document["refresh_token"] != nil || len(document) != 2 {
		t.Fatalf("unexpected grant response: %v", document)
	}
	assertAuthCookieContract(t, response)
	if len(fixture.errors) != 0 {
		t.Fatalf("unexpected reports: %v", fixture.errors)
	}
}

func TestAuthHTTPPortalLoginSelectsPortalApplicationMode(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	fixture.successfulLogin(t, true)
	if fixture.app.loginInput.Mode != AuthHTTPLoginPortal ||
		fixture.app.loginInput.Identifier != "P-100" || fixture.app.loginCalls != 1 {
		t.Fatalf("unexpected portal input: %+v", fixture.app.loginInput)
	}
}

func TestAuthHTTPLoginFailsClosedWithoutIdentityEnumeration(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	fixture.app.loginError = ErrAuthHTTPInvalidCredentials
	first := executeRejectedLogin(t, fixture, "missing-tenant")
	fixture.app.loginError = fmt.Errorf("credential source: %w", ErrAuthHTTPInvalidCredentials)
	second := executeRejectedLogin(t, fixture, "tenant-a")
	if first.Code != http.StatusUnauthorized || second.Code != http.StatusUnauthorized ||
		first.Body.String() != second.Body.String() {
		t.Fatalf("unexpected stable failures: %d/%d", first.Code, second.Code)
	}
	firstBody := decodeAuthResponse(t, first)
	if firstBody["error"].(map[string]any)["code"] != "AUTHENTICATION_FAILED" ||
		strings.Contains(first.Body.String(), "tenant") {
		t.Fatalf("enumerating failure: %s", first.Body.String())
	}
}

func TestAuthHTTPLoginRejectsOriginAndAmbiguousJSONBeforeApplication(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	for _, body := range []string{
		`{"tenant_slug":"tenant-a","email":"a@b.test","password":"secret","extra":true}`,
		`{"tenant_slug":"tenant-a","email":"a@b.test","email":"b@b.test","password":"secret"}`,
	} {
		request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Origin", fixture.origin)
		response := httptest.NewRecorder()
		fixture.handler.login(response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("strict login status=%d body=%s", response.Code, response.Body.String())
		}
	}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(
		`{"tenant_slug":"tenant-a","email":"a@b.test","password":"secret"}`,
	))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "https://attacker.invalid")
	response := httptest.NewRecorder()
	fixture.handler.login(response, request)
	if response.Code != http.StatusForbidden || fixture.app.loginCalls != 0 {
		t.Fatalf("origin was not rejected before application: %d/%d", response.Code, fixture.app.loginCalls)
	}
}

func executeRejectedLogin(
	t *testing.T,
	fixture *authHTTPTestFixture,
	tenant string,
) *httptest.ResponseRecorder {
	t.Helper()
	body := `{"tenant_slug":"` + tenant + `","email":"user@example.com","password":"wrong-password"}`
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", fixture.origin)
	response := httptest.NewRecorder()
	fixture.handler.login(response, request)
	return response
}

func allCleared(value []byte) bool {
	return len(value) > 0 && bytes.Equal(value, make([]byte, len(value)))
}

func assertAuthCookieContract(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	cookies := response.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("cookie count=%d", len(cookies))
	}
	for _, cookie := range cookies {
		if !cookie.Secure || cookie.Domain != "" || cookie.SameSite != http.SameSiteLaxMode {
			t.Fatalf("unsafe cookie: %+v", cookie)
		}
		if cookie.Name == csrf.RefreshCookieName && (!cookie.HttpOnly || cookie.Path != "/api/auth") {
			t.Fatalf("unsafe refresh cookie: %+v", cookie)
		}
		if cookie.Name == csrf.CSRFCookieName && (cookie.HttpOnly || cookie.Path != "/") {
			t.Fatalf("unsafe CSRF cookie: %+v", cookie)
		}
	}
}
