package iam

import (
	"context"
	"encoding/json"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type identityNameProfileAliases struct {
	activeID    identityfacade.IdentityID
	historical  identityfacade.IdentityID
	schemaID    identityfacade.IdentityID
	publicNames map[identityfacade.IdentityID]string
	pinnedNames map[string]identityfacade.IdentityID
	activeCalls int
	pinnedCalls int
}

func (aliases *identityNameProfileAliases) Load(
	context.Context, platformtx.TenantScope,
	identityfacade.ProfileResource, identityfacade.IdentityID,
) (identityfacade.ProfileDocument, error) {
	return identityfacade.ProfileDocument{}, nil
}

func (aliases *identityNameProfileAliases) PrepareNew(
	context.Context, platformtx.TenantScope,
	identityfacade.ProfileResource, []byte,
) (identityfacade.ProfileDocument, error) {
	return identityfacade.ProfileDocument{}, nil
}

func (aliases *identityNameProfileAliases) PreparePatch(
	context.Context, platformtx.TenantScope,
	identityfacade.ProfileResource, identityfacade.IdentityID, []byte,
) (identityfacade.ProfileDocument, error) {
	return identityfacade.ProfileDocument{}, nil
}

func (aliases *identityNameProfileAliases) FieldIDs(
	_ context.Context, _ platformtx.TenantScope,
	_ identityfacade.ProfileResource, _ []string,
) (map[string]identityfacade.IdentityID, error) {
	aliases.activeCalls++
	return map[string]identityfacade.IdentityID{
		identityDisplayNameProfileKey: aliases.activeID,
	}, nil
}

func (aliases *identityNameProfileAliases) FieldIDsForSchema(
	_ context.Context, _ platformtx.TenantScope,
	_ identityfacade.ProfileResource, schemaID identityfacade.IdentityID,
	_ []string,
) (map[string]identityfacade.IdentityID, error) {
	aliases.pinnedCalls++
	aliases.schemaID = schemaID
	if aliases.pinnedNames != nil {
		return aliases.pinnedNames, nil
	}
	return map[string]identityfacade.IdentityID{
		identityDisplayNameProfileKey: aliases.historical,
	}, nil
}

func (aliases *identityNameProfileAliases) PublicFieldNamesForSchema(
	context.Context,
	platformtx.TenantScope,
	identityfacade.ProfileResource,
	identityfacade.IdentityID,
) (map[identityfacade.IdentityID]string, error) {
	return aliases.publicNames, nil
}

func TestIdentityOrganizationAndRoleCreateUseActiveNameProjection(t *testing.T) {
	for _, resource := range []struct {
		public  publicresource.Resource
		profile identityfacade.ProfileResource
	}{
		{publicresource.Organizations, identityfacade.ProfileOrganization},
		{publicresource.Roles, identityfacade.ProfileRole},
	} {
		t.Run(string(resource.public), func(t *testing.T) {
			fixture, port := identityNameProfileFixture(t)
			values := identitySystemResourceValues{
				resource: resource.public,
				values:   map[publicresource.ID]json.RawMessage{},
			}
			profile, err := port.createNameProfile(
				identitySystemResourceCommand{}, resource.profile, values, "Canonical",
			)
			if err != nil || fixture.activeCalls != 1 || fixture.pinnedCalls != 0 ||
				!identityProfileHasValue(profile, fixture.activeID, "Canonical") {
				t.Fatalf("create projection = %s calls=%d/%d err=%v",
					profile, fixture.activeCalls, fixture.pinnedCalls, err)
			}
		})
	}
}

func TestIdentityNameUpdatePinsHistoricalSchemaAndPreservesPatch(t *testing.T) {
	for _, resource := range []struct {
		public  publicresource.Resource
		profile identityfacade.ProfileResource
	}{
		{publicresource.Organizations, identityfacade.ProfileOrganization},
		{publicresource.Roles, identityfacade.ProfileRole},
	} {
		t.Run(string(resource.public), func(t *testing.T) {
			fixture, port := identityNameProfileFixture(t)
			customID := mustIdentitySystemResourceID(
				t, identityCustomProfileTestField,
			)
			values := identitySystemResourceValues{
				resource: resource.public,
				values: map[publicresource.ID]json.RawMessage{
					customID: json.RawMessage(`"retained"`),
				},
			}
			schemaID := mustIdentityProfileID(
				t, "41000000-0000-4000-8000-000000000061",
			)
			profile, err := port.updateNameProfile(
				identitySystemResourceCommand{}, resource.profile,
				schemaID, values, "Renamed",
			)
			if err != nil || fixture.schemaID != schemaID ||
				fixture.pinnedCalls != 1 ||
				!identityProfileHasValue(profile, fixture.historical, "Renamed") ||
				string(identityProfileValue(profile, customID.String())) != `"retained"` {
				t.Fatalf("update projection = %s schema=%v calls=%d err=%v",
					profile, fixture.schemaID, fixture.pinnedCalls, err)
			}
		})
	}
}

func identityNameProfileFixture(
	t *testing.T,
) (*identityNameProfileAliases, *identitySystemResourcePort) {
	t.Helper()
	aliases := &identityNameProfileAliases{
		activeID: mustIdentityProfileID(t, identityDisplayNameTestField),
		historical: mustIdentityProfileID(
			t, "43000000-0000-4000-8000-000000000061",
		),
	}
	return aliases, &identitySystemResourcePort{
		application: &identityHTTPApplication{
			identities: &IdentityRuntime{profiles: aliases},
		},
	}
}

func identityProfileHasValue(
	document []byte,
	id identityfacade.IdentityID,
	expected string,
) bool {
	var values map[string]json.RawMessage
	if json.Unmarshal(document, &values) != nil {
		return false
	}
	var actual string
	return json.Unmarshal(
		values[identityfacadeIDString(id)], &actual,
	) == nil && actual == expected
}

func identityProfileValue(document []byte, key string) json.RawMessage {
	var values map[string]json.RawMessage
	_ = json.Unmarshal(document, &values)
	return values[key]
}

func identityfacadeIDString(id identityfacade.IdentityID) string {
	return publicresource.ID(id).String()
}
