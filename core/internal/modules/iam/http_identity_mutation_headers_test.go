package iam

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestIdentityHTTPCreateRequiresIdempotencyBeforeApplication(t *testing.T) {
	tests := []struct{ path, body string }{
		{"/api/users", `{"email":"user@example.com","password":"long enough password",` +
			`"first_name":"Ada","last_name":"Lovelace","organization_id":` +
			`"67000000-0000-4000-8000-000000000001","status":"ACTIVE"}`},
		{"/api/roles", `{"name":"Operator","is_default":false}`},
		{"/api/organizations", `{"name":"Research","slug":"research"}`},
	}
	for _, test := range tests {
		fixture := newIdentityHTTPTestFixture(t)
		response := fixture.serve(fixture.request(http.MethodPost, test.path, test.body))
		if response.Code != http.StatusBadRequest || len(fixture.app.calls) != 0 {
			t.Fatalf("path=%s status=%d calls=%d", test.path, response.Code, len(fixture.app.calls))
		}
	}
}

func TestIdentityHTTPRevisionedMutationsRequireBothHeaders(t *testing.T) {
	id := "68000000-0000-4000-8000-000000000001"
	tests := []struct{ method, path, body string }{
		{http.MethodPut, "/api/users/" + id, `{"first_name":"Grace"}`},
		{http.MethodDelete, "/api/users/" + id, ""},
		{http.MethodPatch, "/api/users/" + id + "/state", `{"state":"INACTIVE"}`},
		{http.MethodPut, "/api/roles/" + id, `{"name":"Operator"}`},
		{http.MethodDelete, "/api/roles/" + id, ""},
		{http.MethodPatch, "/api/roles/" + id + "/state", `{"state":"RETIRED"}`},
		{http.MethodPut, "/api/organizations/" + id, `{"name":"Research"}`},
		{http.MethodDelete, "/api/organizations/" + id, ""},
		{http.MethodPatch, "/api/organizations/" + id + "/state", `{"state":"SUSPENDED"}`},
	}
	for _, test := range tests {
		for _, missing := range []string{"Idempotency-Key", "If-Match"} {
			fixture := newIdentityHTTPTestFixture(t)
			request := fixture.request(test.method, test.path, test.body)
			identityHTTPMutationHeaders(request)
			request.Header.Del(missing)
			response := fixture.serve(request)
			if response.Code != http.StatusPreconditionRequired || len(fixture.app.calls) != 0 {
				t.Fatalf("%s %s missing=%s status=%d calls=%d", test.method, test.path,
					missing, response.Code, len(fixture.app.calls))
			}
		}
	}
}

func TestIdentityHTTPDeletePassesExactMutationContract(t *testing.T) {
	id := uuid.MustParse("69000000-0000-4000-8000-000000000001")
	tests := []struct{ path, operation string }{
		{"/api/users/" + id.String(), "DeleteUser"},
		{"/api/roles/" + id.String(), "DeleteRole"},
		{"/api/organizations/" + id.String(), "DeleteOrganization"},
	}
	for _, test := range tests {
		fixture := newIdentityHTTPTestFixture(t)
		request := fixture.request(http.MethodDelete, test.path, "")
		identityHTTPMutationHeaders(request)
		response := fixture.serve(request)
		if response.Code != http.StatusOK || len(fixture.app.calls) != 1 ||
			fixture.app.calls[0].operation != test.operation {
			t.Fatalf("path=%s status=%d calls=%v body=%s", test.path,
				response.Code, fixture.app.calls, response.Body.String())
		}
		input := fixture.app.calls[0].input.(IdentityMutation)
		if input.ID != id || input.ExpectedRevision != 1 ||
			input.IdempotencyKey != identityHTTPTestIdempotencyKey {
			t.Fatalf("mutation=%+v", input)
		}
	}
}
