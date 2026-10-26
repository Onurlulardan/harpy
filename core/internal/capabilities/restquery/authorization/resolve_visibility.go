package authorization

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func resolvePolicyVisibilities(
	set requirementSet,
	evaluations []policyEvaluation,
) ([]resolved.ResourceVisibility, error) {
	if !set.validate() || len(evaluations) < 1 {
		return nil, errPolicyEvaluationContract
	}
	result := make([]resolved.ResourceVisibility, 0, len(set.resources))
	seen := make(map[[sha256.Size]byte]struct{}, len(set.resources))
	for _, evaluation := range evaluations {
		if !evaluation.validate() {
			return nil, errPolicyEvaluationContract
		}
		for _, visibility := range evaluation.decision.visibilities {
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
	}
	if len(seen) != len(set.resources) {
		return nil, errPolicyEvaluationContract
	}
	return result, nil
}

func resolvePolicyVisibility(
	resource resourceRequirement,
	visibility Visibility,
) (resolved.ResourceVisibility, error) {
	if !validResourceRequirement(resource.requirement.id, resource) ||
		visibility.Validate() != nil ||
		visibility.resourceRequirement != resource.requirement.id {
		return resolved.ResourceVisibility{}, errPolicyEvaluationContract
	}
	reference, err := resource.catalog.Reference()
	if err != nil {
		return resolved.ResourceVisibility{}, errPolicyEvaluationContract
	}
	switch visibility.mode {
	case VisibilityAll:
		return resolved.NewAllResourceVisibility(
			visibility.resourceRequirement, reference, resource.path,
		)
	case VisibilityNone:
		return resolved.NewNoResourceVisibility(
			visibility.resourceRequirement, reference, resource.path,
		)
	case VisibilityPredicate:
		filter, err := resolveMandatoryFilter(resource.catalog, *visibility.filter)
		if err != nil {
			return resolved.ResourceVisibility{}, err
		}
		return resolved.NewPredicateResourceVisibility(
			visibility.resourceRequirement, reference, resource.path, filter,
		)
	default:
		return resolved.ResourceVisibility{}, errPolicyEvaluationContract
	}
}
