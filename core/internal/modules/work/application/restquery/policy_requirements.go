package restquery

import (
	"crypto/sha256"

	restauthorization "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/authorization"
)

func (policy *workQueryPolicy) matchRequirements(
	requirements []restauthorization.Requirement, resource string,
) (ResolvedCatalog, [sha256.Size]byte, error) {
	catalog := policy.catalogs[resource]
	if catalog.Validate() != nil {
		return ResolvedCatalog{}, [sha256.Size]byte{}, ErrHTTPUnavailable
	}
	var resourceID [sha256.Size]byte
	for _, requirement := range requirements {
		reference, referenceErr := requirement.Catalog()
		kind, kindErr := requirement.Kind()
		identifier, identifierErr := requirement.ID()
		if referenceErr != nil || kindErr != nil || identifierErr != nil ||
			!catalog.matches(reference) || !workQueryRequirementKind(kind) {
			return ResolvedCatalog{}, [sha256.Size]byte{}, ErrHTTPUnavailable
		}
		if kind == restauthorization.RequirementResourceRead {
			if resourceID != [sha256.Size]byte{} {
				return ResolvedCatalog{}, [sha256.Size]byte{}, ErrHTTPUnavailable
			}
			resourceID = identifier
		}
	}
	if resourceID == [sha256.Size]byte{} {
		return ResolvedCatalog{}, [sha256.Size]byte{}, ErrHTTPUnavailable
	}
	return catalog, resourceID, nil
}

func workQueryRequirementKind(kind restauthorization.RequirementKind) bool {
	switch kind {
	case restauthorization.RequirementResourceRead,
		restauthorization.RequirementFieldSelect,
		restauthorization.RequirementFieldFilter,
		restauthorization.RequirementFieldSort,
		restauthorization.RequirementFieldSearch:
		return true
	default:
		return false
	}
}

func workQueryResults(requirements []restauthorization.Requirement,
	allowed bool,
) ([]restauthorization.DecisionResult, error) {
	results := make([]restauthorization.DecisionResult, len(requirements))
	for index, requirement := range requirements {
		identifier, err := requirement.ID()
		if err != nil {
			return nil, ErrHTTPUnavailable
		}
		result, err := restauthorization.NewDecisionResult(identifier, allowed)
		if err != nil {
			return nil, ErrHTTPUnavailable
		}
		results[index] = result
	}
	return results, nil
}
