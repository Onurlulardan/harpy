package iam

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type initialTestChallenge struct {
	cookie, csrf []byte
	expiresAt    time.Time
	revealed     bool
	destroyed    bool
}

func (challenge *initialTestChallenge) ExpiresAt() time.Time {
	return challenge.expiresAt
}

func (challenge *initialTestChallenge) RevealOnce(
	consume func(cookie, csrf []byte) error,
) error {
	if challenge.revealed {
		return ErrAuthHTTPUnavailable
	}
	challenge.revealed = true
	return consume(challenge.cookie, challenge.csrf)
}

func (challenge *initialTestChallenge) Destroy() {
	challenge.destroyed = true
	clear(challenge.cookie)
	clear(challenge.csrf)
}

func TestAuthHTTPTemporaryPasswordReturnsOnlyForcedChangeContinuation(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	cookie, csrf := initialTestWires()
	challenge := &initialTestChallenge{
		cookie: cookie, csrf: csrf, expiresAt: fixture.now.Add(10 * time.Minute),
	}
	fixture.app.initialStartError = nil
	fixture.app.initialStartChallenge = challenge
	request := initialLoginRequest(fixture.origin)
	response := httptest.NewRecorder()
	fixture.handler.login(response, request)
	if response.Code != http.StatusOK || fixture.app.initialStartCalls != 1 ||
		fixture.app.loginCalls != 0 || !challenge.revealed || !challenge.destroyed {
		t.Fatalf("unexpected forced-change start: status=%d calls=%d/%d",
			response.Code, fixture.app.initialStartCalls, fixture.app.loginCalls)
	}
	document := decodeAuthResponse(t, response)
	if document["password_change_required"] != true || document["token"] != nil ||
		len(document) != 2 {
		t.Fatalf("temporary password issued a normal session: %v", document)
	}
	assertInitialStartCookies(t, response)
}

func TestAuthHTTPInitialPasswordCompletesIntoNormalSession(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	fixture.app.initialCompleteGrant = fixture.newGrant(t, fixture.claims)
	cookie, csrf := initialTestWires()
	request := initialPasswordRequest(fixture.origin, cookie, csrf, csrf)
	response := httptest.NewRecorder()
	fixture.handler.initialPassword(response, request)
	if response.Code != http.StatusOK || fixture.app.initialCompleteCalls != 1 {
		t.Fatalf("completion status=%d calls=%d body=%s",
			response.Code, fixture.app.initialCompleteCalls, response.Body.String())
	}
	document := decodeAuthResponse(t, response)
	if document["token"] == "" || document["expires_at"] == "" || len(document) != 2 {
		t.Fatalf("unexpected completion grant: %v", document)
	}
	assertInitialCompletionCookies(t, response)
	if !allCleared(fixture.app.initialCompleteInput.NewPassword) {
		t.Fatal("new password material was not cleared")
	}
}

func TestAuthHTTPInitialPasswordRejectsMismatchedCSRFBeforeApplication(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	cookie, csrf := initialTestWires()
	wrong := append([]byte(nil), csrf...)
	wrong[0] = 'Z'
	request := initialPasswordRequest(fixture.origin, cookie, csrf, wrong)
	response := httptest.NewRecorder()
	fixture.handler.initialPassword(response, request)
	if response.Code != http.StatusForbidden || fixture.app.initialCompleteCalls != 0 {
		t.Fatalf("CSRF mismatch status=%d calls=%d",
			response.Code, fixture.app.initialCompleteCalls)
	}
}

func TestAuthHTTPInitialPasswordReportsPolicyRejectionWithoutSession(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	fixture.app.initialCompleteError = ErrAuthHTTPPasswordRejected
	cookie, csrf := initialTestWires()
	request := initialPasswordRequest(fixture.origin, cookie, csrf, csrf)
	response := httptest.NewRecorder()
	fixture.handler.initialPassword(response, request)
	document := decodeAuthResponse(t, response)
	errorDocument := document["error"].(map[string]any)
	if response.Code != http.StatusBadRequest ||
		errorDocument["code"] != "PASSWORD_POLICY_REJECTED" ||
		len(response.Header().Values("Set-Cookie")) != 0 {
		t.Fatalf("policy response status=%d body=%s", response.Code, response.Body.String())
	}
}

func initialTestWires() ([]byte, []byte) {
	selector := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{1}, 16))
	secret := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{2}, 32))
	csrf := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{3}, 32))
	return []byte(selector + "." + secret), []byte(csrf)
}

func initialLoginRequest(origin string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(
		`{"tenant_slug":"tenant-a","email":"admin@example.com","password":"temporary"}`,
	))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", origin)
	return request
}

func initialPasswordRequest(
	origin string, cookie, csrfCookie, csrfHeader []byte,
) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "/api/auth/initial-password", strings.NewReader(
		`{"tenant_slug":"tenant-a","new_password":"a-strong-new-password"}`,
	))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", origin)
	request.Header.Set("X-CSRF-Token", string(csrfHeader))
	request.AddCookie(&http.Cookie{Name: initialPasswordCookieName, Value: string(cookie)})
	request.AddCookie(&http.Cookie{Name: initialPasswordCSRFName, Value: string(csrfCookie)})
	return request
}

func assertInitialStartCookies(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	found := map[string]*http.Cookie{}
	for _, cookie := range response.Result().Cookies() {
		found[cookie.Name] = cookie
	}
	change, csrf := found[initialPasswordCookieName], found[initialPasswordCSRFName]
	if len(found) != 4 || change == nil || csrf == nil || !change.HttpOnly || csrf.HttpOnly ||
		change.Path != initialPasswordCookiePath || csrf.Path != "/" ||
		!change.Secure || !csrf.Secure || change.SameSite != http.SameSiteStrictMode ||
		csrf.SameSite != http.SameSiteStrictMode || change.MaxAge != 600 || csrf.MaxAge != 600 {
		t.Fatalf("invalid initial cookies: %+v", found)
	}
}

func assertInitialCompletionCookies(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	found := map[string]*http.Cookie{}
	for _, cookie := range response.Result().Cookies() {
		found[cookie.Name] = cookie
	}
	if len(found) != 4 || found[initialPasswordCookieName].MaxAge != -1 ||
		found[initialPasswordCSRFName].MaxAge != -1 {
		t.Fatalf("initial cookies were not retired: %+v", found)
	}
}
