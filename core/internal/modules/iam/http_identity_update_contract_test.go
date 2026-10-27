package iam

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestIdentityHTTPAcceptsV1UpdateRequestFixtures(t *testing.T) {
	id := uuid.MustParse("64000000-0000-4000-8000-000000000001")
	tests := []struct {
		name, path, body, operation string
		result                      func(*identityHTTPTestFixture) any
		assert                      func(*testing.T, any)
	}{
		{
			name: "user legacy status", path: "/api/users/" + id.String(),
			body: `{"status":"INACTIVE"}`, operation: "UpdateUser",
			result: func(f *identityHTTPTestFixture) any {
				view := identityHTTPTestUser(id, f.auth.now)
				view.Status = "INACTIVE"
				return view
			},
			assert: func(t *testing.T, value any) {
				input := value.(IdentityUserUpdate)
				if input.LegacyState != "INACTIVE" || input.ID != id {
					t.Fatalf("user update=%+v", input)
				}
			},
		},
		{
			name: "role", path: "/api/roles/" + id.String(),
			body: `{"description":"Updated V1 role"}`, operation: "UpdateRole",
			result: func(f *identityHTTPTestFixture) any {
				return identityHTTPTestRole(id, f.auth.claims.Tenant(), f.auth.now)
			},
			assert: func(t *testing.T, value any) {
				input := value.(IdentityRoleUpdate)
				if input.Description == nil || *input.Description != "Updated V1 role" {
					t.Fatalf("role update=%+v", input)
				}
			},
		},
		{
			name: "organization", path: "/api/organizations/" + id.String(),
			body: `{"name":"Research Updated"}`, operation: "UpdateOrganization",
			result: func(f *identityHTTPTestFixture) any {
				return identityHTTPTestOrganization(id, f.auth.now)
			},
			assert: func(t *testing.T, value any) {
				input := value.(IdentityOrganizationUpdate)
				if input.Name == nil || *input.Name != "Research Updated" {
					t.Fatalf("organization update=%+v", input)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newIdentityHTTPTestFixture(t)
			fixture.app.results[test.operation] = test.result(fixture)
			request := fixture.request(http.MethodPut, test.path, test.body)
			identityHTTPMutationHeaders(request)
			response := fixture.serve(request)
			if response.Code != http.StatusOK || len(fixture.app.calls) != 1 {
				t.Fatalf("status=%d calls=%d body=%s", response.Code,
					len(fixture.app.calls), response.Body.String())
			}
			call := fixture.app.calls[0]
			if call.operation != test.operation {
				t.Fatalf("operation=%s want=%s", call.operation, test.operation)
			}
			test.assert(t, call.input)
		})
	}
}
