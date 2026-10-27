package iam

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestIdentityHTTPAcceptsCanonicalOrganizationOwners(t *testing.T) {
	fixture := newIdentityHTTPTestFixture(t)
	organizationID := uuid.MustParse("67000000-0000-4000-8000-000000000001")
	ownerA := uuid.MustParse("67000000-0000-4000-8000-000000000002")
	ownerB := uuid.MustParse("67000000-0000-4000-8000-000000000003")
	fixture.app.results["CreateOrganization"] = identityHTTPTestOrganization(
		organizationID, fixture.auth.now,
	)
	request := fixture.request(http.MethodPost, "/api/organizations",
		`{"name":"Research","slug":"research","owner_ids":["`+
			ownerB.String()+`","`+ownerA.String()+`"]}`)
	request.Header.Set("Idempotency-Key", identityHTTPTestIdempotencyKey)
	response := fixture.serve(request)
	if response.Code != http.StatusCreated || len(fixture.app.calls) != 1 {
		t.Fatalf("status=%d calls=%d body=%s", response.Code,
			len(fixture.app.calls), response.Body.String())
	}
	input := fixture.app.calls[0].input.(IdentityOrganizationCreate)
	if len(input.OwnerIDs) != 2 || input.OwnerIDs[0] != ownerA || input.OwnerIDs[1] != ownerB {
		t.Fatalf("owners=%v", input.OwnerIDs)
	}
}

func TestIdentityHTTPRejectsInvalidOrganizationOwnerSets(t *testing.T) {
	owner := "67000000-0000-4000-8000-000000000002"
	tests := []string{
		`{"name":"Research","slug":"research"}`,
		`{"name":"Research","slug":"research","owner_ids":[]}`,
		`{"name":"Research","slug":"research","owner_ids":["` + owner + `","` + owner + `"]}`,
		`{"name":"Research","slug":"research","owner_id":"` + owner +
			`","owner_ids":["` + owner + `"]}`,
	}
	for _, body := range tests {
		fixture := newIdentityHTTPTestFixture(t)
		request := fixture.request(http.MethodPost, "/api/organizations", body)
		request.Header.Set("Idempotency-Key", identityHTTPTestIdempotencyKey)
		response := fixture.serve(request)
		if response.Code != http.StatusBadRequest || len(fixture.app.calls) != 0 {
			t.Fatalf("body=%s status=%d calls=%d", body, response.Code, len(fixture.app.calls))
		}
	}
}

func TestIdentityHTTPAcceptsLegacyOwnerIDOnOrganizationUpdate(t *testing.T) {
	fixture := newIdentityHTTPTestFixture(t)
	id := uuid.MustParse("67000000-0000-4000-8000-000000000010")
	owner := uuid.MustParse("67000000-0000-4000-8000-000000000011")
	fixture.app.results["UpdateOrganization"] = identityHTTPTestOrganization(id, fixture.auth.now)
	request := fixture.request(http.MethodPut, "/api/organizations/"+id.String(),
		`{"owner_id":"`+owner.String()+`"}`)
	identityHTTPMutationHeaders(request)
	response := fixture.serve(request)
	if response.Code != http.StatusOK || len(fixture.app.calls) != 1 {
		t.Fatalf("status=%d calls=%d body=%s", response.Code,
			len(fixture.app.calls), response.Body.String())
	}
	input := fixture.app.calls[0].input.(IdentityOrganizationUpdate)
	if input.OwnerIDs == nil || len(*input.OwnerIDs) != 1 || (*input.OwnerIDs)[0] != owner {
		t.Fatalf("owners=%v", input.OwnerIDs)
	}
}

func TestIdentityHTTPRejectsNullOrAmbiguousOrganizationOwnerUpdate(t *testing.T) {
	id := uuid.MustParse("67000000-0000-4000-8000-000000000020")
	owner := "67000000-0000-4000-8000-000000000021"
	for _, body := range []string{
		`{"name":"Changed","owner_ids":null}`,
		`{"owner_id":"` + owner + `","owner_ids":["` + owner + `"]}`,
		`{"owner_ids":[]}`,
	} {
		fixture := newIdentityHTTPTestFixture(t)
		request := fixture.request(http.MethodPut, "/api/organizations/"+id.String(), body)
		identityHTTPMutationHeaders(request)
		response := fixture.serve(request)
		if response.Code != http.StatusBadRequest || len(fixture.app.calls) != 0 {
			t.Fatalf("body=%s status=%d calls=%d", body, response.Code, len(fixture.app.calls))
		}
	}
}
