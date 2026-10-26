package iam

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	iamapplication "github.com/Onurlulardan/harpy/core/internal/modules/iam/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/csrf"
	"github.com/google/uuid"
)

func (fixture *authHTTPTestFixture) sessionView(
	claims accesstoken.Claims,
) AuthHTTPSession {
	portalID, _ := claims.PortalBinding()
	return AuthHTTPSession{
		UserID: claims.Subject(), TenantID: claims.Tenant(),
		SessionID: claims.Session(), TenantSlug: "tenant-a",
		Email: "admin@example.com", DisplayName: "Instance Owner",
		ClientID: claims.Client(), AccessMode: claims.AccessMode(),
		PortalBindingID: portalID, TokenVersion: claims.TokenVersion(),
		AuthorizationVersion:       claims.AuthorizationVersion(),
		PasswordCredentialRevision: pointerInt64(3),
		PlatformAuthorizationClock: claims.PlatformAuthorizationClock(),
		SecurityEpoch:              claims.SecurityEpoch(),
		DeploymentMode:             iamapplication.DeploymentModeOnPremise,
		TenantPurpose:              iamapplication.TenantPurposeOnPremisePrimary,
		PlatformCapabilities:       []string{},
		TenantCapabilities:         []string{"forms.read", "users.read"},
	}
}

func pointerInt64(value int64) *int64 { return &value }

func (fixture *authHTTPTestFixture) successfulLogin(
	t *testing.T,
	portal bool,
) (*httptest.ResponseRecorder, *http.Cookie, *http.Cookie) {
	t.Helper()
	claims := fixture.claims
	path := "/api/auth/login"
	body := `{"tenant_slug":"tenant-a","email":"admin@example.com","password":"valid-password"}`
	if portal {
		claims = fixture.newClaims(t, true)
		path = "/api/auth/portal/login"
		body = `{"tenant_slug":"tenant-a","identifier":"P-100","password":"valid-password"}`
	}
	fixture.app.loginGrant = fixture.newGrant(t, claims)
	request := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", fixture.origin)
	response := httptest.NewRecorder()
	if portal {
		fixture.handler.portalLogin(response, request)
	} else {
		fixture.handler.login(response, request)
	}
	if response.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", response.Code, response.Body.String())
	}
	var refreshCookie, csrfCookie *http.Cookie
	for _, cookie := range response.Result().Cookies() {
		switch cookie.Name {
		case csrf.RefreshCookieName:
			refreshCookie = cookie
		case csrf.CSRFCookieName:
			csrfCookie = cookie
		}
	}
	if refreshCookie == nil || csrfCookie == nil {
		t.Fatalf("auth cookies are missing: %v", response.Header().Values("Set-Cookie"))
	}
	return response, refreshCookie, csrfCookie
}

func (fixture *authHTTPTestFixture) browserRequest(
	method string,
	path string,
	refreshCookie *http.Cookie,
	csrfCookie *http.Cookie,
) *http.Request {
	var body io.Reader = http.NoBody
	if method == http.MethodPost {
		body = ioBody(`{}`)
	}
	request := httptest.NewRequest(method, path, body)
	request.Header.Set("Origin", fixture.origin)
	if method == http.MethodPost {
		request.Header.Set("Content-Type", "application/json")
	}
	if refreshCookie != nil {
		request.AddCookie(refreshCookie)
	}
	if csrfCookie != nil {
		request.AddCookie(csrfCookie)
		request.Header.Set(csrf.HeaderName, csrfCookie.Value)
	}
	return request
}

func ioBody(value string) io.Reader {
	return bytes.NewReader([]byte(value))
}

func decodeAuthResponse(t *testing.T, response *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var document map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &document); err != nil {
		t.Fatal(err)
	}
	return document
}

func assertCanonicalUUID(t *testing.T, value any, want uuid.UUID) {
	t.Helper()
	if value != want.String() {
		t.Fatalf("UUID=%v want=%s", value, want)
	}
}
