package iam

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestIdentityUserViewPublishesSingularOrganizationOnlyForExactOne(t *testing.T) {
	first := uuid.MustParse("8a000000-0000-4000-8000-000000000001")
	second := uuid.MustParse("8a000000-0000-4000-8000-000000000002")
	tests := []struct {
		name string
		ids  []uuid.UUID
		want *uuid.UUID
	}{
		{name: "none", ids: []uuid.UUID{}},
		{name: "single", ids: []uuid.UUID{first}, want: &first},
		{name: "multiple", ids: []uuid.UUID{first, second}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := identitySingleOrganizationID(test.ids)
			if test.want == nil && got != nil ||
				test.want != nil && (got == nil || *got != *test.want) {
				t.Fatalf("alias=%v want=%v", got, test.want)
			}
		})
	}
}

func TestIdentityHTTPUserRecordEmitsTruthfulOrganizationAlias(t *testing.T) {
	organizationID := uuid.MustParse("8b000000-0000-4000-8000-000000000001")
	fixture := newIdentityHTTPTestFixture(t)
	view := identityHTTPTestUser(
		uuid.MustParse("8b000000-0000-4000-8000-000000000002"),
		fixture.auth.now,
	)
	view.OrganizationIDs = []uuid.UUID{organizationID}
	view.OrganizationID = identitySingleOrganizationID(view.OrganizationIDs)
	fixture.app.results["GetUser"] = view

	response := fixture.serve(fixture.request(
		http.MethodGet, "/api/users/"+view.ID.String(), "",
	))
	var envelope struct {
		Data map[string]json.RawMessage `json:"data"`
	}
	if response.Code != http.StatusOK ||
		json.Unmarshal(response.Body.Bytes(), &envelope) != nil ||
		string(envelope.Data["organization_id"]) != `"`+organizationID.String()+`"` {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestIdentityUserViewRejectsAmbiguousSingularOrganizationAlias(t *testing.T) {
	view := identityHTTPTestUser(
		uuid.MustParse("8c000000-0000-4000-8000-000000000001"),
		newIdentityHTTPTestFixture(t).auth.now,
	)
	first := uuid.MustParse("8c000000-0000-4000-8000-000000000002")
	second := uuid.MustParse("8c000000-0000-4000-8000-000000000003")
	view.OrganizationIDs = []uuid.UUID{first, second}
	view.OrganizationID = &first
	if validIdentityUserView(view) {
		t.Fatal("multi-organization view accepted a fabricated singular alias")
	}
}
