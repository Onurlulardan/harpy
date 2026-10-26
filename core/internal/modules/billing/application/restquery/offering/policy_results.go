package offering

import (
	"crypto/sha256"

	restauthorization "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/authorization"
)

func policyResults(
	catalogs *Catalog,
	requirements []restauthorization.Requirement,
	allowed bool,
) ([]restauthorization.DecisionResult, [sha256.Size]byte, error) {
	results := make([]restauthorization.DecisionResult, len(requirements))
	var resourceRequirement [sha256.Size]byte
	for index, requirement := range requirements {
		reference, referenceErr := requirement.Catalog()
		kind, kindErr := requirement.Kind()
		identity, identityErr := requirement.ID()
		if referenceErr != nil || kindErr != nil || identityErr != nil ||
			!catalogs.Matches(reference) || !offeringRequirement(kind) {
			return nil, [sha256.Size]byte{}, ErrContract
		}
		if kind == restauthorization.RequirementResourceRead {
			if resourceRequirement != [sha256.Size]byte{} {
				return nil, [sha256.Size]byte{}, ErrContract
			}
			resourceRequirement = identity
		}
		result, err := restauthorization.NewDecisionResult(identity, allowed)
		if err != nil {
			return nil, [sha256.Size]byte{}, ErrContract
		}
		results[index] = result
	}
	if resourceRequirement == [sha256.Size]byte{} {
		return nil, [sha256.Size]byte{}, ErrContract
	}
	return results, resourceRequirement, nil
}

func offeringRequirement(kind restauthorization.RequirementKind) bool {
	switch kind {
	case restauthorization.RequirementResourceRead,
		restauthorization.RequirementFieldSelect,
		restauthorization.RequirementFieldFilter,
		restauthorization.RequirementFieldSort,
		restauthorization.RequirementFieldSearch,
		restauthorization.RequirementExactCount:
		return true
	default:
		return false
	}
}
