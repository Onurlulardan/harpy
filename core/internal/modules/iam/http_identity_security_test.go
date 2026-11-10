package iam

import (
	"net/http"
	"testing"
)

func TestIdentityHTTPRequiresVerifiedOperatorAndMatchingTenant(t *testing.T) {
	tests := map[string]func(*identityHTTPTestFixture, *http.Request, *testing.T){
		"missing bearer": func(_ *identityHTTPTestFixture, request *http.Request, _ *testing.T) {
			request.Header.Del("Authorization")
		},
		"duplicate bearer": func(_ *identityHTTPTestFixture, request *http.Request, _ *testing.T) {
			request.Header.Add("Authorization", request.Header.Get("Authorization"))
		},
		"foreign tenant": func(_ *identityHTTPTestFixture, request *http.Request, _ *testing.T) {
			request.Header.Set("X-Tenant-ID", "90000000-0000-4000-8000-000000000001")
		},
		"duplicate tenant": func(fixture *identityHTTPTestFixture, request *http.Request, _ *testing.T) {
			request.Header.Add("X-Tenant-ID", fixture.auth.claims.Tenant().String())
			request.Header.Add("X-Tenant-ID", fixture.auth.claims.Tenant().String())
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			fixture := newIdentityHTTPTestFixture(t)
			fixture.app.results["Query"] = IdentityQueryView{Items: nil}
			request := fixture.request(http.MethodGet, "/api/users", "")
			mutate(fixture, request, t)
			response := fixture.serve(request)
			if response.Code != http.StatusUnauthorized || len(fixture.app.calls) != 0 {
				t.Fatalf("status=%d calls=%d body=%s", response.Code, len(fixture.app.calls), response.Body.String())
			}
		})
	}
}

func TestIdentityHTTPTenantSlugNeverEstablishesAuthority(t *testing.T) {
	fixture := newIdentityHTTPTestFixture(t)
	fixture.app.results["Query"] = IdentityQueryView{Items: nil}
	request := fixture.request(http.MethodGet, "/api/users", "")
	request.Header.Set("X-Tenant-ID", fixture.auth.claims.Tenant().String())
	request.Header.Set("X-Tenant-Slug", "foreign-looking-but-non-authoritative")
	response := fixture.serve(request)
	if response.Code != http.StatusOK || len(fixture.app.calls) != 1 {
		t.Fatalf("status=%d calls=%d body=%s", response.Code, len(fixture.app.calls), response.Body.String())
	}
	if fixture.app.calls[0].claims.Tenant() != fixture.auth.claims.Tenant() {
		t.Fatal("application did not receive the verified token tenant")
	}
}
