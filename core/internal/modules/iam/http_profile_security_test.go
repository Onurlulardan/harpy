package iam

import (
	"net/http"
	"testing"
)

func TestProfileHTTPRequiresVerifiedOperatorAndMatchingTenant(t *testing.T) {
	tests := map[string]func(*profileHTTPTestFixture, *http.Request, *testing.T){
		"missing bearer": func(_ *profileHTTPTestFixture, request *http.Request, _ *testing.T) {
			request.Header.Del("Authorization")
		},
		"duplicate bearer": func(_ *profileHTTPTestFixture, request *http.Request, _ *testing.T) {
			request.Header.Add("Authorization", request.Header.Get("Authorization"))
		},
		"foreign tenant": func(_ *profileHTTPTestFixture, request *http.Request, _ *testing.T) {
			request.Header.Set("X-Tenant-ID", "90000000-0000-4000-8000-000000000001")
		},
		"portal bearer": func(fixture *profileHTTPTestFixture, request *http.Request, t *testing.T) {
			claims := fixture.auth.newClaims(t, true)
			access, err := fixture.auth.security.accessSigner.Sign(claims)
			if err != nil {
				t.Fatal(err)
			}
			request.Header.Set("Authorization", "Bearer "+access)
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			fixture := newProfileHTTPTestFixture(t)
			fixture.app.results["GetSelf"] = profileHTTPTestSelfView()
			request := fixture.request(http.MethodGet, "/api/profile", "")
			mutate(fixture, request, t)
			response := fixture.serve(request)
			if response.Code != http.StatusUnauthorized || len(fixture.app.calls) != 0 {
				t.Fatalf("status=%d calls=%d body=%s", response.Code, len(fixture.app.calls), response.Body)
			}
		})
	}
}

func TestProfileHTTPTenantSlugIsNotAuthority(t *testing.T) {
	fixture := newProfileHTTPTestFixture(t)
	fixture.app.results["GetSelf"] = profileHTTPTestSelfView()
	request := fixture.request(http.MethodGet, "/api/profile", "")
	request.Header.Set("X-Tenant-ID", fixture.auth.claims.Tenant().String())
	request.Header.Set("X-Tenant-Slug", "non-authoritative-slug")
	response := fixture.serve(request)
	if response.Code != http.StatusOK || len(fixture.app.calls) != 1 ||
		fixture.app.calls[0].claims.Tenant() != fixture.auth.claims.Tenant() {
		t.Fatalf("status=%d calls=%+v", response.Code, fixture.app.calls)
	}
}
