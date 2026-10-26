package authorization

import (
	"crypto/sha256"
	"testing"
)

func exactDecision(
	t *testing.T,
	request PolicyRequest,
	denied map[[sha256.Size]byte]bool,
) PolicyDecision {
	t.Helper()
	results := make([]DecisionResult, 0, len(request.requirements))
	visibilities := make([]Visibility, 0)
	for _, requirement := range request.requirements {
		allowed := !denied[requirement.id]
		result, err := NewDecisionResult(requirement.id, allowed)
		if err != nil {
			t.Fatal(err)
		}
		results = append(results, result)
		if allowed && requirement.kind == RequirementResourceRead {
			visibility, visibilityErr := NewAllVisibility(requirement.id)
			if visibilityErr != nil {
				t.Fatal(visibilityErr)
			}
			visibilities = append(visibilities, visibility)
		}
	}
	decision, err := NewPolicyDecision(
		request.binding,
		sha256.Sum256([]byte("owner-context-revision")),
		results,
		visibilities,
	)
	if err != nil {
		t.Fatal(err)
	}
	return decision
}
