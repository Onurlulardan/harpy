package authorization

import (
	"crypto/sha256"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

var (
	authorizedPlanTenant = uuid.MustParse("71000000-0000-4000-8000-000000000001")
	authorizedPlanActor  = uuid.MustParse("72000000-0000-4000-8000-000000000001")
)

func allowAllPlanFixture(
	t *testing.T,
	spec resolved.QuerySpec,
) (requirementSet, []policyEvaluation) {
	t.Helper()
	set, err := extractRequirements(spec)
	if err != nil {
		t.Fatal(err)
	}
	identity, _ := spec.IdentityHash()
	requests, err := groupPolicyRequests(
		authorizedPlanTenant, authorizedPlanActor, identity, set.requirements,
	)
	if err != nil {
		t.Fatal(err)
	}
	evaluations := make([]policyEvaluation, 0, len(requests))
	for _, request := range requests {
		decision := exactDecision(t, request, nil)
		validated, err := validatePolicyDecision(request, decision)
		if err != nil {
			t.Fatal(err)
		}
		evaluations = append(evaluations, policyEvaluation{
			owner: request.owner, request: request, decision: validated,
		})
	}
	return set, evaluations
}

func evaluationWithVisibility(
	t *testing.T,
	request PolicyRequest,
	build func(Requirement) Visibility,
) policyEvaluation {
	t.Helper()
	results := make([]DecisionResult, 0, len(request.requirements))
	visibilities := make([]Visibility, 0)
	for _, requirement := range request.requirements {
		result, err := NewDecisionResult(requirement.id, true)
		if err != nil {
			t.Fatal(err)
		}
		results = append(results, result)
		if requirement.kind == RequirementResourceRead {
			visibilities = append(visibilities, build(requirement))
		}
	}
	decision, err := NewPolicyDecision(
		request.binding, sha256.Sum256([]byte("fixture-policy-revision")),
		results, visibilities,
	)
	if err != nil {
		t.Fatal(err)
	}
	validated, err := validatePolicyDecision(request, decision)
	if err != nil {
		t.Fatal(err)
	}
	return policyEvaluation{
		owner: request.owner, request: request, decision: validated,
	}
}
