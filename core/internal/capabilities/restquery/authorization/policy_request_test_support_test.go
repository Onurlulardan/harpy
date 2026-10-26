package authorization

import (
	"crypto/sha256"
	"testing"

	"github.com/google/uuid"
)

func policyRequirements(t *testing.T) []Requirement {
	t.Helper()
	return []Requirement{
		mustRequirement(
			t, RequirementFieldFilter, "", "/where",
			uuid.MustParse(testFieldID), uuid.Nil,
		),
		mustRequirement(
			t, RequirementResourceRead, "", "", uuid.Nil, uuid.Nil,
		),
	}
}

func mustPolicyRequest(t *testing.T) PolicyRequest {
	t.Helper()
	request, err := newPolicyRequest(
		uuid.MustParse("50000000-0000-4000-8000-000000000001"),
		uuid.MustParse("60000000-0000-4000-8000-000000000001"),
		sha256.Sum256([]byte("query")),
		"forms",
		policyRequirements(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	return request
}
