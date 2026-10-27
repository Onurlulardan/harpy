package iam

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func newInvitationHTTPTestFixture(
	t *testing.T,
) (*InvitationHTTP, *invitationHTTPTestApplication, string, uuid.UUID) {
	t.Helper()
	auth := newAuthHTTPTestFixture(t)
	app := &invitationHTTPTestApplication{
		createResult: InvitationView{
			ID: uuid.New(), Email: "invitee@example.com", Role: "Operator",
			InvitedBy: auth.claims.Subject(), Status: "PENDING",
			ExpiresAt: auth.now.Add(time.Hour), CreatedAt: auth.now, Revision: 1,
		},
		verifyResult: InvitationVerifyView{
			Email: "invitee@example.com", Role: "Operator", TenantName: "Acme",
			TenantSlug: "acme", ExpiresAt: auth.now.Add(time.Hour),
		},
		acceptResult: InvitationAcceptView{
			UserID: uuid.New(), TenantSlug: "acme", RequiresLogin: true,
		},
	}
	handler, err := newInvitationHTTP(
		auth.security, app, func(error) {}, func() time.Time { return auth.now },
	)
	if err != nil {
		t.Fatal(err)
	}
	token, err := auth.security.accessSigner.Sign(auth.claims)
	if err != nil {
		t.Fatal(err)
	}
	return handler, app, token, auth.claims.Tenant()
}

func invitationProtectedRequest(
	method, path, body, token string, tenant uuid.UUID,
) *http.Request {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set(authTenantIDHeader, tenant.String())
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	return request
}

func TestInvitationHTTPProtectedRoutesBindExactBearerAndTenant(t *testing.T) {
	handler, app, token, tenant := newInvitationHTTPTestFixture(t)
	valid := invitationProtectedRequest(http.MethodGet, "/api/invitations", "", token, tenant)
	response := httptest.NewRecorder()
	handler.list(response, valid)
	if response.Code != http.StatusOK || app.listCalls != 1 {
		t.Fatalf("valid request=%d calls=%d", response.Code, app.listCalls)
	}
	wrong := invitationProtectedRequest(http.MethodGet, "/api/invitations", "", token, uuid.New())
	response = httptest.NewRecorder()
	handler.list(response, wrong)
	if response.Code != http.StatusUnauthorized || app.listCalls != 1 {
		t.Fatalf("tenant substitution=%d calls=%d", response.Code, app.listCalls)
	}
	duplicate := invitationProtectedRequest(http.MethodGet, "/api/invitations", "", token, tenant)
	duplicate.Header.Add("Authorization", "Bearer "+token)
	response = httptest.NewRecorder()
	handler.list(response, duplicate)
	if response.Code != http.StatusUnauthorized || app.listCalls != 1 {
		t.Fatalf("duplicate bearer=%d calls=%d", response.Code, app.listCalls)
	}
}

func TestInvitationHTTPCreateRequiresStrictJSONAndIdempotency(t *testing.T) {
	handler, app, token, tenant := newInvitationHTTPTestFixture(t)
	body := `{"email":"invitee@example.com","role":"Operator"}`
	request := invitationProtectedRequest(http.MethodPost, "/api/invitations", body, token, tenant)
	response := httptest.NewRecorder()
	handler.create(response, request)
	if response.Code != http.StatusBadRequest || app.createCalls != 0 {
		t.Fatalf("missing idempotency=%d calls=%d", response.Code, app.createCalls)
	}
	request = invitationProtectedRequest(http.MethodPost, "/api/invitations", strings.TrimSuffix(body, "}")+`,"token":"forbidden"}`, token, tenant)
	request.Header.Set("Idempotency-Key", "invitation-create-0001")
	response = httptest.NewRecorder()
	handler.create(response, request)
	if response.Code != http.StatusBadRequest || app.createCalls != 0 {
		t.Fatalf("unknown JSON=%d calls=%d", response.Code, app.createCalls)
	}
}

func TestInvitationHTTPPublicCredentialSurfacesRedactSecrets(t *testing.T) {
	handler, app, _, _ := newInvitationHTTPTestFixture(t)
	secret := "fciv1_selector_secret"
	request := httptest.NewRequest(http.MethodGet, "/api/invitations/verify?token="+secret+"&tenant=acme", nil)
	response := httptest.NewRecorder()
	handler.verify(response, request)
	if response.Code != http.StatusOK || app.verifyCalls != 1 ||
		string(app.verifyInput.Token) != secret || request.URL.RawQuery != "" ||
		strings.Contains(request.RequestURI, secret) || bytes.Contains(response.Body.Bytes(), []byte(secret)) {
		t.Fatalf("verify did not redact credential: status=%d uri=%q body=%s", response.Code, request.RequestURI, response.Body.Bytes())
	}
	body := `{"token":"` + secret + `","first_name":"Ada","last_name":"Lovelace","password":"Correct-Horse-99!","tenant_slug":"acme","custom_attributes":{}}`
	request = httptest.NewRequest(http.MethodPost, "/api/invitations/accept", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	handler.accept(response, request)
	responseLower := bytes.ToLower(response.Body.Bytes())
	if response.Code != http.StatusCreated || app.acceptCalls != 1 ||
		bytes.Contains(responseLower, []byte("token")) ||
		bytes.Contains(responseLower, []byte("password")) ||
		bytes.Contains(responseLower, []byte("session")) ||
		!bytes.Contains(responseLower, []byte(`"requires_login":true`)) {
		t.Fatalf("accept response contract failed: status=%d body=%s", response.Code, response.Body.Bytes())
	}
}
