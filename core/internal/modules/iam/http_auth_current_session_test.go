package iam

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestAuthHTTPCurrentSessionReturnsCanonicalCurrentAuthority(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	token, err := fixture.security.accessSigner.Sign(fixture.claims)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/auth/session", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set(authTenantIDHeader, fixture.claims.Tenant().String())
	request.Header.Set(authTenantSlugHeader, "tenant-a")
	response := httptest.NewRecorder()
	fixture.handler.currentSession(response, request)
	if response.Code != http.StatusOK || fixture.app.currentCalls != 1 {
		t.Fatalf("session status/calls=%d/%d body=%s", response.Code, fixture.app.currentCalls, response.Body.String())
	}
	document := decodeAuthResponse(t, response)
	assertCanonicalUUID(t, document["session_id"], fixture.claims.Session())
	assertCanonicalUUID(t, document["user_id"], fixture.claims.Subject())
	assertCanonicalUUID(t, document["tenant_id"], fixture.claims.Tenant())
	if document["tenant_slug"] != "tenant-a" ||
		document["email"] != "admin@example.com" ||
		document["display_name"] != "Instance Owner" ||
		document["access_mode"] != "operator_console" ||
		document["client_id"] != "harpy-operator-console-v1" ||
		document["deployment_mode"] != "onpremise" ||
		document["platform_capabilities"] == nil ||
		document["tenant_capabilities"] == nil ||
		document["auth_time"] == nil || document["expires_at"] == nil ||
		document["password_credential_revision"] != float64(3) ||
		document["portal_binding_id"] != nil || len(document) != 18 {
		t.Fatalf("unexpected current session: %v", document)
	}
}

func TestAuthHTTPCurrentSessionEmitsPasswordlessNullRevision(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	fixture.app.currentSession.PasswordCredentialRevision = nil
	token, err := fixture.security.accessSigner.Sign(fixture.claims)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/auth/session", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	fixture.handler.currentSession(response, request)
	document := decodeAuthResponse(t, response)
	value, present := document["password_credential_revision"]
	if response.Code != http.StatusOK || !present || value != nil {
		t.Fatalf("passwordless revision status/present/value=%d/%t/%v", response.Code, present, value)
	}
}

func TestAuthHTTPCurrentSessionRejectsTenantAndVersionSubstitution(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	token, err := fixture.security.accessSigner.Sign(fixture.claims)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/auth/session", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set(authTenantIDHeader, uuid.NewString())
	response := httptest.NewRecorder()
	fixture.handler.currentSession(response, request)
	if response.Code != http.StatusUnauthorized || fixture.app.currentCalls != 0 {
		t.Fatalf("tenant substitution status/calls=%d/%d", response.Code, fixture.app.currentCalls)
	}
	fixture.app.currentSession.TokenVersion++
	request = httptest.NewRequest(http.MethodGet, "/api/auth/session", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response = httptest.NewRecorder()
	fixture.handler.currentSession(response, request)
	if response.Code != http.StatusUnauthorized || fixture.app.currentCalls != 1 {
		t.Fatalf("version substitution status/calls=%d/%d", response.Code, fixture.app.currentCalls)
	}
}

func TestAuthHTTPCurrentPortalSessionRequiresExactPortalBinding(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	claims := fixture.newClaims(t, true)
	token, err := fixture.security.accessSigner.Sign(claims)
	if err != nil {
		t.Fatal(err)
	}
	fixture.app.currentSession = fixture.sessionView(claims)
	request := httptest.NewRequest(http.MethodGet, "/api/auth/session", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	fixture.handler.currentSession(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("portal session status=%d body=%s", response.Code, response.Body.String())
	}
	document := decodeAuthResponse(t, response)
	if document["portal_binding_id"] == nil || document["access_mode"] != "portal" {
		t.Fatalf("portal binding missing: %v", document)
	}
	fixture.app.currentSession.PortalBindingID = uuid.New()
	response = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/auth/session", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	fixture.handler.currentSession(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("portal substitution status=%d", response.Code)
	}
}

func TestAuthHTTPCurrentSessionRejectsAmbiguousBearer(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	request := httptest.NewRequest(http.MethodGet, "/api/auth/session", nil)
	request.Header.Add("Authorization", "Bearer invalid")
	request.Header.Add("Authorization", "Bearer another")
	response := httptest.NewRecorder()
	fixture.handler.currentSession(response, request)
	if response.Code != http.StatusUnauthorized || fixture.app.currentCalls != 0 {
		t.Fatalf("ambiguous bearer status/calls=%d/%d", response.Code, fixture.app.currentCalls)
	}
}
