package authorization

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

func validateAuthorizedPlanInputs(
	spec resolved.QuerySpec,
	set requirementSet,
	tenantID uuid.UUID,
	actorID uuid.UUID,
	evaluations []policyEvaluation,
) (requirementSet, error) {
	if spec.Validate() != nil || !set.validate() ||
		!validAuthorizationUUID(tenantID) || !validAuthorizationUUID(actorID) {
		return requirementSet{}, errPolicyEvaluationContract
	}
	extracted, err := extractRequirements(spec)
	if err != nil || !sameRequirementSets(set, extracted) {
		return requirementSet{}, errPolicyEvaluationContract
	}
	identity, err := spec.IdentityHash()
	if err != nil {
		return requirementSet{}, errPolicyEvaluationContract
	}
	expected, err := groupPolicyRequests(
		tenantID, actorID, identity, extracted.requirements,
	)
	if err != nil || len(expected) != len(evaluations) {
		return requirementSet{}, errPolicyEvaluationContract
	}
	for index, evaluation := range evaluations {
		if !evaluation.validate() || evaluation.owner != expected[index].owner ||
			!sameDigest(evaluation.request.binding, expected[index].binding) {
			return requirementSet{}, errPolicyEvaluationContract
		}
	}
	return extracted, nil
}

func sameRequirementSets(left, right requirementSet) bool {
	if !left.validate() || !right.validate() ||
		len(left.requirements) != len(right.requirements) ||
		len(left.resources) != len(right.resources) {
		return false
	}
	expected := make(map[[sha256.Size]byte]struct{}, len(right.requirements))
	for _, requirement := range right.requirements {
		expected[requirement.id] = struct{}{}
	}
	for _, requirement := range left.requirements {
		if _, found := expected[requirement.id]; !found {
			return false
		}
		delete(expected, requirement.id)
	}
	return len(expected) == 0
}
