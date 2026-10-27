package iam

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestLifecycleHTTPProtectedRouteRejectsBearerAndTenantSubstitution(t *testing.T) {
	auth := newAuthHTTPTestFixture(t)
	app := &lifecycleHTTPTestApplication{}
	handler, err := newLifecycleHTTP(
		auth.security, app, func(error) {}, func() time.Time { return auth.now },
	)
	if err != nil {
		t.Fatal(err)
	}
	token, err := auth.security.accessSigner.Sign(auth.claims)
	if err != nil {
		t.Fatal(err)
	}
	request := lifecycleProtectedRequest(token, auth.claims.Tenant())
	response := httptest.NewRecorder()
	handler.listAuthenticators(response, request)
	if response.Code != http.StatusOK || app.listAuthenticatorCalls != 1 {
		t.Fatalf("valid request status=%d calls=%d", response.Code, app.listAuthenticatorCalls)
	}
	request = lifecycleProtectedRequest(token, uuid.New())
	response = httptest.NewRecorder()
	handler.listAuthenticators(response, request)
	if response.Code != http.StatusUnauthorized || app.listAuthenticatorCalls != 1 {
		t.Fatalf("tenant substitution status=%d calls=%d", response.Code, app.listAuthenticatorCalls)
	}
	request = lifecycleProtectedRequest(token, auth.claims.Tenant())
	request.Header.Add("Authorization", "Bearer "+token)
	response = httptest.NewRecorder()
	handler.listAuthenticators(response, request)
	if response.Code != http.StatusUnauthorized || app.listAuthenticatorCalls != 1 {
		t.Fatalf("duplicate bearer status=%d calls=%d", response.Code, app.listAuthenticatorCalls)
	}
}

func lifecycleProtectedRequest(token string, tenant uuid.UUID) *http.Request {
	request := httptest.NewRequest(http.MethodGet, "/api/iam/authenticators", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set(authTenantIDHeader, tenant.String())
	return request
}
