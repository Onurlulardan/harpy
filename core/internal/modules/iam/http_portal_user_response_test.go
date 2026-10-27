package iam

import (
	"bytes"
	"net/http"
	"strings"
	"testing"
)

func TestPortalUserResponseIsRevisionedAndSecretFree(t *testing.T) {
	fixture := newPortalUserHTTPTestFixture(t)
	body := `{"subject_id":"018f47f4-7b1b-7000-8000-000000000012",` +
		`"subject_type":"employee","portal_profile_form_slug":"employees"}`
	request := fixture.request(http.MethodPost, "/api/auth/portal-users", body)
	request.Header.Set("Idempotency-Key", portalUserHTTPTestKey)
	response := fixture.serve(request)
	lower := bytes.ToLower(response.Body.Bytes())
	if response.Code != http.StatusCreated || response.Header().Get("ETag") != `"revision-3"` ||
		!bytes.Contains(lower, []byte(`"revision":3`)) ||
		bytes.Contains(lower, []byte("setup_token")) || bytes.Contains(lower, []byte("password")) ||
		bytes.Contains(lower, []byte("setup_url")) {
		t.Fatalf("status=%d etag=%q body=%s", response.Code,
			response.Header().Get("ETag"), response.Body.String())
	}
}

func TestPortalUserHTTPRejectsPortalBearerAndDuplicateAuthority(t *testing.T) {
	fixture := newPortalUserHTTPTestFixture(t)
	portalClaims := fixture.auth.newClaims(t, true)
	portalToken, err := fixture.auth.security.accessSigner.Sign(portalClaims)
	if err != nil {
		t.Fatal(err)
	}
	request := fixture.request(http.MethodPost, "/api/auth/portal-users", `{}`)
	request.Header.Set("Authorization", "Bearer "+portalToken)
	request.Header.Set("Idempotency-Key", portalUserHTTPTestKey)
	response := fixture.serve(request)
	if response.Code != http.StatusUnauthorized || len(fixture.app.calls) != 0 {
		t.Fatalf("portal bearer status=%d calls=%d", response.Code, len(fixture.app.calls))
	}
	request = fixture.request(http.MethodPost, "/api/auth/portal-users", `{}`)
	request.Header.Set("Idempotency-Key", portalUserHTTPTestKey)
	request.Header.Add("Authorization", "Bearer "+fixture.token)
	response = fixture.serve(request)
	if response.Code != http.StatusUnauthorized || len(fixture.app.calls) != 0 ||
		strings.Contains(response.Body.String(), fixture.token) {
		t.Fatalf("duplicate bearer status=%d calls=%d", response.Code, len(fixture.app.calls))
	}
}
