package iam

import (
	"testing"

	"github.com/google/uuid"
)

func TestIdentityRoleScopeValidationUsesVerifiedTenantOrExactOrganization(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	tenant := fixture.claims.Tenant()
	organization := uuid.MustParse("86000000-0000-4000-8000-000000000001")
	base := IdentityRoleCreate{IdempotencyKey: "stable-key", Name: "Operator"}
	tests := []struct {
		name  string
		input IdentityRoleCreate
		valid bool
	}{
		{"tenant", mergeIdentityRoleScope(base, "TENANT", tenant, uuid.Nil), true},
		{"organization", mergeIdentityRoleScope(base, "ORGANIZATION", organization, organization), true},
		{"foreign tenant", mergeIdentityRoleScope(base, "TENANT", organization, uuid.Nil), false},
		{"tenant with organization", mergeIdentityRoleScope(base, "TENANT", tenant, organization), false},
		{"organization mismatch", mergeIdentityRoleScope(base, "ORGANIZATION", organization, tenant), false},
		{"unknown", mergeIdentityRoleScope(base, "GLOBAL", tenant, uuid.Nil), false},
		{"missing key", mergeIdentityRoleScope(IdentityRoleCreate{Name: "Operator"}, "TENANT", tenant, uuid.Nil), false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := identityRoleScopeValid(fixture.claims, test.input); actual != test.valid {
				t.Fatalf("valid=%t want=%t input=%+v", actual, test.valid, test.input)
			}
		})
	}
}

func mergeIdentityRoleScope(
	input IdentityRoleCreate, kind string, scopeID uuid.UUID, organizationID uuid.UUID,
) IdentityRoleCreate {
	input.ScopeKind = kind
	input.ScopeID = scopeID
	input.OrganizationID = organizationID
	return input
}
