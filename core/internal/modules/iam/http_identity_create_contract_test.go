package iam

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestIdentityHTTPAcceptsV1CreateRequestFixtures(t *testing.T) {
	organizationID := uuid.MustParse("61000000-0000-4000-8000-000000000001")
	roleID := uuid.MustParse("62000000-0000-4000-8000-000000000001")
	userID := uuid.MustParse("63000000-0000-4000-8000-000000000001")
	tests := []struct {
		name, path, body, operation string
		result                      func(*identityHTTPTestFixture) any
		assert                      func(*testing.T, identityHTTPTestCall, *identityHTTPTestFixture)
	}{
		{
			name: "user", path: "/api/users", operation: "CreateUser",
			body: `{"email":"user@example.com","password":"correct horse battery staple",` +
				`"first_name":"Ada","last_name":"Lovelace","organization_id":"` +
				organizationID.String() + `","role_ids":["` + roleID.String() + `"],"status":"ACTIVE"}`,
			result: func(f *identityHTTPTestFixture) any { return identityHTTPTestUser(userID, f.auth.now) },
			assert: func(t *testing.T, call identityHTTPTestCall, _ *identityHTTPTestFixture) {
				input := call.input.(IdentityUserCreate)
				if input.OrganizationID != organizationID || len(input.RoleIDs) != 1 ||
					input.RoleIDs[0] != roleID || input.IdempotencyKey != identityHTTPTestIdempotencyKey {
					t.Fatalf("user input=%+v", input)
				}
				for _, value := range input.Password {
					if value != 0 {
						t.Fatal("password was not wiped after application call")
					}
				}
			},
		},
		{
			name: "role legacy scope inference", path: "/api/roles", operation: "CreateRole",
			body: `{"name":"Operator","description":"V1 role","is_default":false,` +
				`"organization_id":"` + organizationID.String() + `"}`,
			result: func(f *identityHTTPTestFixture) any {
				view := identityHTTPTestRole(roleID, f.auth.claims.Tenant(), f.auth.now)
				view.ScopeKind, view.ScopeID, view.OrganizationID = "ORGANIZATION", organizationID, organizationID
				return view
			},
			assert: func(t *testing.T, call identityHTTPTestCall, _ *identityHTTPTestFixture) {
				input := call.input.(IdentityRoleCreate)
				if input.ScopeKind != "ORGANIZATION" || input.ScopeID != organizationID ||
					input.OrganizationID != organizationID {
					t.Fatalf("role input=%+v", input)
				}
			},
		},
		{
			name: "organization", path: "/api/organizations", operation: "CreateOrganization",
			body: `{"name":"Research","slug":"research","owner_id":"` + userID.String() + `"}`,
			result: func(f *identityHTTPTestFixture) any {
				return identityHTTPTestOrganization(organizationID, f.auth.now)
			},
			assert: func(t *testing.T, call identityHTTPTestCall, _ *identityHTTPTestFixture) {
				input := call.input.(IdentityOrganizationCreate)
				if input.Name != "Research" || input.Slug != "research" ||
					len(input.OwnerIDs) != 1 || input.OwnerIDs[0] != userID {
					t.Fatalf("organization input=%+v", input)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newIdentityHTTPTestFixture(t)
			fixture.app.results[test.operation] = test.result(fixture)
			request := fixture.request(http.MethodPost, test.path, test.body)
			request.Header.Set("Idempotency-Key", identityHTTPTestIdempotencyKey)
			response := fixture.serve(request)
			if response.Code != http.StatusCreated || response.Header().Get("ETag") != `"revision-2"` ||
				len(fixture.app.calls) != 1 || fixture.app.calls[0].operation != test.operation {
				t.Fatalf("status=%d etag=%q calls=%v body=%s", response.Code,
					response.Header().Get("ETag"), fixture.app.calls, response.Body.String())
			}
			test.assert(t, fixture.app.calls[0], fixture)
		})
	}
}
