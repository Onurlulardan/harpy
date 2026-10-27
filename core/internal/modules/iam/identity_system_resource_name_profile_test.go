package iam

import (
	"encoding/json"
	"errors"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	"github.com/google/uuid"
)

const identityDisplayNameTestField = "43000000-0000-4000-8000-000000000001"
const identityCustomProfileTestField = "43000000-0000-4000-8000-000000000002"

func TestIdentityNameProfileProjectsDisplayAndPreservesCustomPatch(t *testing.T) {
	displayID := mustIdentitySystemResourceID(t, identityDisplayNameTestField)
	customID := mustIdentitySystemResourceID(t, identityCustomProfileTestField)
	values := identitySystemResourceValues{
		resource: publicresource.Organizations,
		values: map[publicresource.ID]json.RawMessage{
			mustIdentitySystemResourceID(
				t, publicresource.OrganizationNameFieldID,
			): json.RawMessage(`"Acme"`),
			customID: json.RawMessage(`"retained"`),
		},
	}
	profile, err := values.profileWithDisplayName(
		mustIdentityProfileID(t, identityDisplayNameTestField), "Acme",
	)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]json.RawMessage
	if json.Unmarshal(profile, &document) != nil ||
		string(document[displayID.String()]) != `"Acme"` ||
		string(document[customID.String()]) != `"retained"` ||
		len(document) != 2 {
		t.Fatalf("projected profile = %s", profile)
	}
}

func TestIdentityNameProfileRejectsMismatchedDualInput(t *testing.T) {
	displayID := mustIdentitySystemResourceID(t, identityDisplayNameTestField)
	values := identitySystemResourceValues{
		resource: publicresource.Roles,
		values: map[publicresource.ID]json.RawMessage{
			displayID: json.RawMessage(`"Different"`),
		},
	}
	_, err := values.profileWithDisplayName(
		mustIdentityProfileID(t, identityDisplayNameTestField), "Canonical",
	)
	if !errors.Is(err, publicresource.ErrContract) {
		t.Fatalf("mismatched dual input error = %v", err)
	}
}

func TestIdentityNameProfileAcceptsMatchingDualInput(t *testing.T) {
	displayID := mustIdentitySystemResourceID(t, identityDisplayNameTestField)
	values := identitySystemResourceValues{
		resource: publicresource.Roles,
		values: map[publicresource.ID]json.RawMessage{
			displayID: json.RawMessage(`"Canonical"`),
		},
	}
	profile, err := values.profileWithDisplayName(
		mustIdentityProfileID(t, identityDisplayNameTestField), "Canonical",
	)
	if err != nil || string(identityProfileValue(profile, displayID.String())) !=
		`"Canonical"` {
		t.Fatalf("matching dual input = %s/%v", profile, err)
	}
}

func TestIdentityUserProfileFlowRemainsUnprojected(t *testing.T) {
	customID := mustIdentitySystemResourceID(t, identityCustomProfileTestField)
	values := identitySystemResourceValues{
		resource: publicresource.Users,
		values: map[publicresource.ID]json.RawMessage{
			customID: json.RawMessage(`"Ada"`),
		},
	}
	profile, err := values.profile()
	if err != nil || string(profile) !=
		`{"43000000-0000-4000-8000-000000000002":"Ada"}` {
		t.Fatalf("user profile = %s/%v", profile, err)
	}
}

func mustIdentityProfileID(
	t *testing.T,
	raw string,
) identityfacade.IdentityID {
	t.Helper()
	id, err := uuid.Parse(raw)
	if err != nil || id == uuid.Nil {
		t.Fatal(err)
	}
	return identityfacade.IdentityID(id)
}
