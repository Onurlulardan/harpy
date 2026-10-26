package iam

import (
	"net/http"
	"testing"
)

func TestPortalUserCreateAndInviteRequireIdempotency(t *testing.T) {
	identifier := "018f47f4-7b1b-7000-8000-000000000012"
	tests := []struct{ path, body string }{
		{"/api/auth/portal-users", `{"subject_id":"` + identifier +
			`","subject_type":"employee","portal_profile_form_slug":"employees"}`},
		{"/api/auth/portal-users/" + identifier + "/invite", ""},
	}
	for _, test := range tests {
		fixture := newPortalUserHTTPTestFixture(t)
		response := fixture.serve(fixture.request(http.MethodPost, test.path, test.body))
		if response.Code != http.StatusBadRequest || len(fixture.app.calls) != 0 {
			t.Fatalf("path=%s status=%d calls=%d", test.path, response.Code,
				len(fixture.app.calls))
		}
	}
}

func TestPortalUserRevisionedMutationsRequireBothHeaders(t *testing.T) {
	identifier := "018f47f4-7b1b-7000-8000-000000000010"
	tests := []struct{ method, path, body string }{
		{http.MethodPatch, "/api/auth/portal-users/" + identifier + "/status",
			`{"status":"INACTIVE"}`},
		{http.MethodPost, "/api/auth/portal-users/" + identifier + "/console-access", ""},
		{http.MethodDelete, "/api/auth/portal-users/" + identifier + "/console-access", ""},
	}
	for _, test := range tests {
		for _, missing := range []string{"Idempotency-Key", "If-Match"} {
			fixture := newPortalUserHTTPTestFixture(t)
			request := fixture.request(test.method, test.path, test.body)
			portalUserMutationHeaders(request, 3)
			request.Header.Del(missing)
			response := fixture.serve(request)
			if response.Code != http.StatusPreconditionRequired || len(fixture.app.calls) != 0 {
				t.Fatalf("%s missing=%s status=%d calls=%d", test.path, missing,
					response.Code, len(fixture.app.calls))
			}
		}
	}
}

func TestPortalUserStatusPassesExactRevisionContract(t *testing.T) {
	fixture := newPortalUserHTTPTestFixture(t)
	identifier := "018f47f4-7b1b-7000-8000-000000000010"
	request := fixture.request(http.MethodPatch,
		"/api/auth/portal-users/"+identifier+"/status", `{"status":"inactive"}`)
	portalUserMutationHeaders(request, 3)
	response := fixture.serve(request)
	if response.Code != http.StatusOK || len(fixture.app.calls) != 1 {
		t.Fatalf("status=%d calls=%d body=%s", response.Code,
			len(fixture.app.calls), response.Body.String())
	}
	input := fixture.app.calls[0].input
	if input.ID.String() != identifier || input.ExpectedRevision != 3 ||
		input.IdempotencyKey != portalUserHTTPTestKey || input.Status != "INACTIVE" {
		t.Fatalf("input=%+v", input)
	}
}
