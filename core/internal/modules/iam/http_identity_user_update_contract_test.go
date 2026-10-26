package iam

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestIdentityHTTPAcceptsV1UserRelationUpdate(t *testing.T) {
	userID := uuid.MustParse("90000000-0000-4000-8000-000000000001")
	organizationID := uuid.MustParse("90000000-0000-4000-8000-000000000002")
	roleID := uuid.MustParse("90000000-0000-4000-8000-000000000003")
	fixture := newIdentityHTTPTestFixture(t)
	view := identityHTTPTestUser(userID, fixture.auth.now)
	view.OrganizationIDs = []uuid.UUID{organizationID}
	view.OrganizationID = identitySingleOrganizationID(view.OrganizationIDs)
	view.RoleIDs = []uuid.UUID{roleID}
	fixture.app.results["UpdateUser"] = view

	request := fixture.request(
		http.MethodPut, "/api/users/"+userID.String(),
		`{"organization_id":"`+organizationID.String()+`","role_ids":["`+
			roleID.String()+`"]}`,
	)
	identityHTTPMutationHeaders(request)
	response := fixture.serve(request)
	if response.Code != http.StatusOK || len(fixture.app.calls) != 1 ||
		fixture.app.calls[0].operation != "UpdateUser" {
		t.Fatalf("status=%d calls=%v body=%s",
			response.Code, fixture.app.calls, response.Body.String())
	}
	input := fixture.app.calls[0].input.(IdentityUserUpdate)
	if input.OrganizationID == nil || *input.OrganizationID != organizationID ||
		input.RoleIDs == nil || len(*input.RoleIDs) != 1 || (*input.RoleIDs)[0] != roleID {
		t.Fatalf("input=%+v", input)
	}
}

func TestIdentityHTTPTreatsEmptyRoleIDsAsExplicitClear(t *testing.T) {
	userID := uuid.MustParse("91000000-0000-4000-8000-000000000001")
	fixture := newIdentityHTTPTestFixture(t)
	fixture.app.results["UpdateUser"] = identityHTTPTestUser(userID, fixture.auth.now)
	request := fixture.request(
		http.MethodPut, "/api/users/"+userID.String(), `{"role_ids":[]}`,
	)
	identityHTTPMutationHeaders(request)
	response := fixture.serve(request)
	if response.Code != http.StatusOK || len(fixture.app.calls) != 1 {
		t.Fatalf("status=%d calls=%v body=%s",
			response.Code, fixture.app.calls, response.Body.String())
	}
	input := fixture.app.calls[0].input.(IdentityUserUpdate)
	if input.RoleIDs == nil || len(*input.RoleIDs) != 0 {
		t.Fatalf("role clear input=%+v", input)
	}
}

func TestIdentityHTTPDoesNotTreatNullOrganizationAsMembershipClear(t *testing.T) {
	userID := uuid.MustParse("92000000-0000-4000-8000-000000000001")
	fixture := newIdentityHTTPTestFixture(t)
	request := fixture.request(
		http.MethodPut, "/api/users/"+userID.String(), `{"organization_id":null}`,
	)
	identityHTTPMutationHeaders(request)
	response := fixture.serve(request)
	if response.Code != http.StatusBadRequest || len(fixture.app.calls) != 0 {
		t.Fatalf("status=%d calls=%v body=%s",
			response.Code, fixture.app.calls, response.Body.String())
	}
}
