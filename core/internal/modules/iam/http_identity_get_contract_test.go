package iam

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestIdentityHTTPGetUsesCanonicalPathAndETag(t *testing.T) {
	id := uuid.MustParse("87000000-0000-4000-8000-000000000001")
	tests := []struct {
		name, path, operation string
		result                func(*identityHTTPTestFixture) any
	}{
		{
			"user", "/api/users/" + id.String(), "GetUser",
			func(f *identityHTTPTestFixture) any { return identityHTTPTestUser(id, f.auth.now) },
		},
		{
			"role", "/api/roles/" + id.String(), "GetRole",
			func(f *identityHTTPTestFixture) any {
				return identityHTTPTestRole(id, f.auth.claims.Tenant(), f.auth.now)
			},
		},
		{
			"organization", "/api/organizations/" + id.String(), "GetOrganization",
			func(f *identityHTTPTestFixture) any {
				return identityHTTPTestOrganization(id, f.auth.now)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newIdentityHTTPTestFixture(t)
			fixture.app.results[test.operation] = test.result(fixture)
			response := fixture.serve(fixture.request(http.MethodGet, test.path, ""))
			if response.Code != http.StatusOK || response.Header().Get("ETag") != `"revision-2"` ||
				len(fixture.app.calls) != 1 || fixture.app.calls[0].operation != test.operation ||
				fixture.app.calls[0].input.(uuid.UUID) != id {
				t.Fatalf("status=%d etag=%q calls=%v body=%s", response.Code,
					response.Header().Get("ETag"), fixture.app.calls, response.Body.String())
			}
		})
	}
}

func TestIdentityHTTPGetRejectsSubstitutedApplicationRecord(t *testing.T) {
	requested := uuid.MustParse("88000000-0000-4000-8000-000000000001")
	returned := uuid.MustParse("89000000-0000-4000-8000-000000000001")
	fixture := newIdentityHTTPTestFixture(t)
	fixture.app.results["GetUser"] = identityHTTPTestUser(returned, fixture.auth.now)
	response := fixture.serve(fixture.request(
		http.MethodGet, "/api/users/"+requested.String(), "",
	))
	if response.Code != http.StatusServiceUnavailable || len(fixture.errors) != 1 ||
		!responseHasIdentityErrorCode(response.Body.Bytes(), "IAM_UNAVAILABLE") {
		t.Fatalf("status=%d errors=%v body=%s", response.Code,
			fixture.errors, response.Body.String())
	}
}
