package iam

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/google/uuid"
)

func TestIdentityOperationIDIsStableAndAuthorityScoped(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	claims := fixture.claims
	first := identityOperationID(claims, "iam.identity.user.create", "stable-key")
	if first != identityOperationID(claims, "iam.identity.user.create", "stable-key") ||
		first.Version() != 8 {
		t.Fatalf("operation ID is not deterministic UUIDv8: %s", first)
	}
	for name, candidate := range map[string]uuid.UUID{
		"operation": identityOperationID(claims, "iam.identity.role.create", "stable-key"),
		"key":       identityOperationID(claims, "iam.identity.user.create", "other-key"),
		"tenant": identityOperationID(identityHTTPTestClaims(
			t, fixture, claims.Subject(), uuid.MustParse("71000000-0000-4000-8000-000000000001"),
		), "iam.identity.user.create", "stable-key"),
		"actor": identityOperationID(identityHTTPTestClaims(
			t, fixture, uuid.MustParse("72000000-0000-4000-8000-000000000001"), claims.Tenant(),
		), "iam.identity.user.create", "stable-key"),
	} {
		if candidate == first {
			t.Fatalf("%s substitution retained operation ID", name)
		}
	}
}

func TestIdentityUserGraphIDsAreStableAndDistinct(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	first := identityNewUserIDs(fixture.claims, "stable-key")
	second := identityNewUserIDs(fixture.claims, "stable-key")
	if first != second {
		t.Fatalf("graph IDs changed across replay: %+v %+v", first, second)
	}
	seen := map[uuid.UUID]struct{}{}
	for name, id := range map[string]uuid.UUID{
		"operation": first.operation, "user": first.user, "identifier": first.identifier,
		"access": first.accessMode, "default-role": first.defaultAssignment,
	} {
		if id == uuid.Nil {
			t.Fatalf("%s ID is nil", name)
		}
		if _, duplicate := seen[id]; duplicate {
			t.Fatalf("%s ID collides: %s", name, id)
		}
		seen[id] = struct{}{}
	}
}

func identityHTTPTestClaims(
	t *testing.T, fixture *authHTTPTestFixture, subject uuid.UUID, tenant uuid.UUID,
) accesstoken.Claims {
	t.Helper()
	claims, err := accesstoken.NewClaims(fixture.profile, accesstoken.ClaimsInput{
		Subject: subject, Session: uuid.MustParse("73000000-0000-4000-8000-000000000001"),
		Tenant: tenant, Client: accesstoken.ClientOperatorConsole, IssuedAt: fixture.now,
		JTI:          uuid.MustParse("74000000-0000-4000-8000-000000000001"),
		TokenVersion: 3, AuthorizationVersion: 5,
		PlatformAuthorizationClock: 7, SecurityEpoch: [32]byte{1},
		AccessMode: accesstoken.AccessModeOperator, AuthenticationTime: fixture.now,
		Assurance: accesstoken.AssuranceBaseline,
		Methods:   []accesstoken.AuthMethod{accesstoken.AuthMethodPassword},
	})
	if err != nil {
		t.Fatal(err)
	}
	return claims
}
