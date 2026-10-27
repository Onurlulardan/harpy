package iam

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"testing"

	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestIdentityCreateCarriesPublicProfileNames(t *testing.T) {
	fixture := newIdentityHTTPTestFixture(t)
	id := uuid.MustParse("83000000-0000-4000-8000-000000000001")
	ownerID := uuid.MustParse("83000000-0000-4000-8000-000000000002")
	fixture.app.results["CreateOrganization"] = identityHTTPTestOrganization(
		id, fixture.auth.now,
	)
	request := fixture.request(
		http.MethodPost, "/api/organizations",
		`{"name":"Research","slug":"research","owner_ids":["`+
			ownerID.String()+`"],"department_code":"RND"}`,
	)
	request.Header.Set("Idempotency-Key", identityHTTPTestIdempotencyKey)
	response := fixture.serve(request)
	if response.Code != http.StatusCreated || len(fixture.app.calls) != 1 {
		t.Fatalf("status=%d calls=%d body=%s", response.Code,
			len(fixture.app.calls), response.Body.String())
	}
	input, ok := fixture.app.calls[0].input.(IdentityOrganizationCreate)
	var values map[string]json.RawMessage
	if !ok || len(input.OwnerIDs) != 1 || input.OwnerIDs[0] != ownerID ||
		json.Unmarshal(input.CustomAttributes, &values) != nil ||
		string(values["department_code"]) != `"RND"` {
		t.Fatalf("profile input=%s", input.CustomAttributes)
	}
}

func TestIdentityRecordFlattensPublicProfileWithoutUUIDKeys(t *testing.T) {
	view := identityHTTPTestUser(
		uuid.MustParse("83000000-0000-4000-8000-000000000001"),
		newAuthHTTPTestFixture(t).now,
	)
	view.CustomAttributes = json.RawMessage(`{"department_code":"RND"}`)
	record, err := identityFlattenPublicProfile(view)
	if err != nil || string(record["department_code"]) != `"RND"` {
		t.Fatalf("record=%v err=%v", record, err)
	}
	if _, nested := record["custom_attributes"]; nested {
		t.Fatal("public profile remained nested")
	}
	for key := range record {
		if _, parseErr := uuid.Parse(key); parseErr == nil {
			t.Fatalf("stable profile UUID leaked as response key: %s", key)
		}
	}
}

func TestIdentityAccountTypeComesOnlyFromActiveModes(t *testing.T) {
	if identityUserAccountType([]string{"operator_console", "portal"}) != "operator_portal" ||
		identityUserAccountType([]string{"operator_console"}) != "operator" ||
		identityUserAccountType([]string{"portal"}) != "portal" {
		t.Fatal("active access mode compatibility projection drifted")
	}
}

func TestIdentityPublicProfileOmitsRestrictedStableFields(t *testing.T) {
	publicID := mustIdentityProfileID(t, identityCustomProfileTestField)
	restrictedID := mustIdentityProfileID(
		t, "43000000-0000-4000-8000-000000000099",
	)
	schemaID := mustIdentityProfileID(
		t, "41000000-0000-4000-8000-000000000099",
	)
	values := []byte(`{"43000000-0000-4000-8000-000000000002":"public","43000000-0000-4000-8000-000000000099":"secret"}`)
	document, err := identityfacade.NewProfileDocument(
		identityfacade.ProfileUser, schemaID, values, sha256.Sum256(values),
	)
	if err != nil {
		t.Fatal(err)
	}
	aliases := &identityNameProfileAliases{publicNames: map[identityfacade.IdentityID]string{
		publicID: "department_code",
	}}
	application := identityHTTPApplication{identities: &IdentityRuntime{profiles: aliases}}
	actual, err := application.publicProfileValues(
		context.Background(), platformtx.TenantScope{}, identityfacade.ProfileUser, document,
	)
	if err != nil || string(actual["department_code"]) != `"public"` {
		t.Fatalf("profile=%v error=%v", actual, err)
	}
	if _, leaked := actual[uuid.UUID(restrictedID).String()]; leaked {
		t.Fatal("restricted stable field leaked")
	}
}
