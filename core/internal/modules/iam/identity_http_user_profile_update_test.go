package iam

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"testing"

	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestIdentityUserPartialNameUpdateSynchronizesDisplayName(t *testing.T) {
	schemaID := mustIdentityProfileID(
		t, "41000000-0000-4000-8000-000000000088",
	)
	ids := map[string]identityfacade.IdentityID{
		"display_name": mustIdentityProfileID(
			t, "43000000-0000-4000-8000-000000000081",
		),
		"first_name": mustIdentityProfileID(
			t, "43000000-0000-4000-8000-000000000082",
		),
		"last_name": mustIdentityProfileID(
			t, "43000000-0000-4000-8000-000000000083",
		),
		"phone": mustIdentityProfileID(
			t, "43000000-0000-4000-8000-000000000084",
		),
	}
	currentValues := map[string]string{
		uuid.UUID(ids["display_name"]).String(): "Ada Lovelace",
		uuid.UUID(ids["first_name"]).String():   "Ada",
		uuid.UUID(ids["last_name"]).String():    "Lovelace",
	}
	current, _ := json.Marshal(currentValues)
	profile, err := identityfacade.NewProfileDocument(
		identityfacade.ProfileUser, schemaID, current, sha256.Sum256(current),
	)
	if err != nil {
		t.Fatal(err)
	}
	aliases := &identityNameProfileAliases{pinnedNames: ids}
	application := identityHTTPApplication{
		identities: &IdentityRuntime{profiles: aliases},
	}
	firstName := "Grace"
	patch, err := application.userProfileUpdateValues(
		context.Background(), platformtx.TenantScope{},
		identityfacade.User{Profile: profile}, nil, &firstName, nil, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	var values map[string]json.RawMessage
	if json.Unmarshal(patch, &values) != nil ||
		string(values[uuid.UUID(ids["first_name"]).String()]) != `"Grace"` ||
		string(values[uuid.UUID(ids["display_name"]).String()]) != `"Grace Lovelace"` {
		t.Fatalf("partial profile patch=%s", patch)
	}
	if _, rewrote := values[uuid.UUID(ids["last_name"]).String()]; rewrote {
		t.Fatal("partial name patch rewrote the unchanged field")
	}
	if aliases.pinnedCalls != 1 || aliases.schemaID != schemaID {
		t.Fatalf("profile alias lookup was not pinned: %+v", aliases)
	}
}
