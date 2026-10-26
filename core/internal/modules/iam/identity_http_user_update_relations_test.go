package iam

import (
	"context"
	"errors"
	"testing"

	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	"github.com/google/uuid"
)

func TestIdentityUserUpdateRequestHashCanonicalizesRoleSet(t *testing.T) {
	first := uuid.MustParse("8d000000-0000-4000-8000-000000000001")
	second := uuid.MustParse("8d000000-0000-4000-8000-000000000002")
	roles := []uuid.UUID{second, first}
	input := IdentityUserUpdate{
		IdentityMutation: IdentityMutation{
			ID:               uuid.MustParse("8d000000-0000-4000-8000-000000000003"),
			ExpectedRevision: 4,
		},
		RoleIDs: &roles,
	}
	firstDigest, err := identityUserUpdateRequestSHA(input, []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	reordered := []uuid.UUID{first, second}
	input.RoleIDs = &reordered
	secondDigest, err := identityUserUpdateRequestSHA(input, []byte(`{}`))
	if err != nil || secondDigest != firstDigest {
		t.Fatalf("reordered digest=%x want=%x error=%v", secondDigest, firstDigest, err)
	}
	empty := []uuid.UUID{}
	input.RoleIDs = &empty
	emptyDigest, err := identityUserUpdateRequestSHA(input, []byte(`{}`))
	if err != nil || emptyDigest == firstDigest {
		t.Fatal("explicit empty replacement is not bound to the request digest")
	}
}

func TestIdentityLegacyUserRelationSetsApplyExactV1Replacement(t *testing.T) {
	oldOrganization := uuid.MustParse("8e000000-0000-4000-8000-000000000001")
	newOrganization := uuid.MustParse("8e000000-0000-4000-8000-000000000002")
	oldRole := uuid.MustParse("8e000000-0000-4000-8000-000000000003")
	newRole := uuid.MustParse("8e000000-0000-4000-8000-000000000004")
	user := identityfacade.User{
		OrganizationIDs: identityFacadeIDs([]uuid.UUID{oldOrganization}),
		RoleIDs:         identityFacadeIDs([]uuid.UUID{oldRole}),
	}
	replacementRoles := []uuid.UUID{newRole}
	organizations, roles := identityLegacyUserRelationSets(user, IdentityUserUpdate{
		OrganizationID: &newOrganization, RoleIDs: &replacementRoles,
	})
	if len(organizations) != 1 || organizations[0] != newOrganization ||
		len(roles) != 1 || roles[0] != newRole {
		t.Fatalf("organizations=%v roles=%v", organizations, roles)
	}
	roles[0] = oldRole
	if replacementRoles[0] != newRole {
		t.Fatal("relation resolver mutated the request role set")
	}
}

func TestIdentityUserUpdateForbidsSelfRelationMutationBeforeIO(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	organizationID := uuid.MustParse("8f000000-0000-4000-8000-000000000001")
	_, err := (&identityHTTPApplication{}).UpdateUser(
		context.Background(), fixture.claims,
		IdentityUserUpdate{
			IdentityMutation: IdentityMutation{
				ID: fixture.claims.Subject(), ExpectedRevision: 1,
				IdempotencyKey: "stable-key",
			},
			OrganizationID: &organizationID,
		},
	)
	if !errors.Is(err, ErrIdentityHTTPForbidden) {
		t.Fatalf("error=%v", err)
	}
}
