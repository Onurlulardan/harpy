package iam

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestIdentityHTTPAcceptsExactStateRequestFixtures(t *testing.T) {
	id := uuid.MustParse("65000000-0000-4000-8000-000000000001")
	tests := []struct {
		name, path, state, operation string
		result                       func(*identityHTTPTestFixture) any
	}{
		{
			name: "user", path: "/api/users/" + id.String() + "/state",
			state: "INACTIVE", operation: "ChangeUserState",
			result: func(f *identityHTTPTestFixture) any {
				view := identityHTTPTestUser(id, f.auth.now)
				view.Status = "INACTIVE"
				return view
			},
		},
		{
			name: "role", path: "/api/roles/" + id.String() + "/state",
			state: "RETIRED", operation: "ChangeRoleState",
			result: func(f *identityHTTPTestFixture) any {
				view := identityHTTPTestRole(id, f.auth.claims.Tenant(), f.auth.now)
				view.Status = "RETIRED"
				return view
			},
		},
		{
			name: "organization", path: "/api/organizations/" + id.String() + "/state",
			state: "SUSPENDED", operation: "ChangeOrganizationState",
			result: func(f *identityHTTPTestFixture) any {
				view := identityHTTPTestOrganization(id, f.auth.now)
				view.Status = "SUSPENDED"
				return view
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newIdentityHTTPTestFixture(t)
			fixture.app.results[test.operation] = test.result(fixture)
			request := fixture.request(http.MethodPatch, test.path,
				`{"state":"`+test.state+`"}`)
			identityHTTPMutationHeaders(request)
			response := fixture.serve(request)
			if response.Code != http.StatusOK || len(fixture.app.calls) != 1 {
				t.Fatalf("status=%d calls=%d body=%s", response.Code,
					len(fixture.app.calls), response.Body.String())
			}
			call := fixture.app.calls[0]
			input := call.input.(IdentityStateMutation)
			if call.operation != test.operation || input.ID != id || input.State != test.state ||
				input.ExpectedRevision != 1 || input.IdempotencyKey != identityHTTPTestIdempotencyKey {
				t.Fatalf("operation=%s mutation=%+v", call.operation, input)
			}
		})
	}
}
