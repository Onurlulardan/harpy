package authorization

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func resolveGlobalVisibilities(
	set requirementSet,
	decision validatedDecision,
) ([]resolved.ResourceVisibility, error) {
	if !set.validate() || decision.requestBinding == [sha256.Size]byte{} {
		return nil, errPolicyEvaluationContract
	}
	result := make([]resolved.ResourceVisibility, 0, len(set.resources))
	seen := make(map[[sha256.Size]byte]struct{}, len(set.resources))
	for _, visibility := range decision.visibilities {
		identity := visibility.resourceRequirement
		resource, found := set.resources[identity]
		if !found {
			return nil, errPolicyEvaluationContract
		}
		if _, duplicate := seen[identity]; duplicate {
			return nil, errPolicyEvaluationContract
		}
		resolvedVisibility, err := resolvePolicyVisibility(resource, visibility)
		if err != nil {
			return nil, err
		}
		seen[identity] = struct{}{}
		result = append(result, resolvedVisibility)
	}
	if len(seen) != len(set.resources) {
		return nil, errPolicyEvaluationContract
	}
	return result, nil
}
