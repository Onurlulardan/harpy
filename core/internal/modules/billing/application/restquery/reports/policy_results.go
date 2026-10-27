package reports

import (
	"crypto/sha256"

	restauthorization "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/authorization"
)

func policyResults(
	catalogs *Catalog,
	requirements []restauthorization.Requirement,
) (resourceSpec, []restauthorization.DecisionResult, [sha256.Size]byte, error) {
	results := make([]restauthorization.DecisionResult, len(requirements))
	var selected resourceSpec
	var resourceID [sha256.Size]byte
	for _, requirement := range requirements {
		reference, referenceErr := requirement.Catalog()
		kind, kindErr := requirement.Kind()
		identity, identityErr := requirement.ID()
		spec, found := catalogs.resourceFor(reference)
		if referenceErr != nil || kindErr != nil || identityErr != nil || !found ||
			!reportRequirement(spec.Aggregate, kind) {
			return resourceSpec{}, nil, [sha256.Size]byte{}, ErrContract
		}
		if selected.Name == "" {
			selected = spec
		} else if selected.Name != spec.Name {
			return resourceSpec{}, nil, [sha256.Size]byte{}, ErrContract
		}
		if kind == restauthorization.RequirementResourceRead {
			if resourceID != [sha256.Size]byte{} {
				return resourceSpec{}, nil, [sha256.Size]byte{}, ErrContract
			}
			resourceID = identity
		}
	}
	if selected.Name == "" || resourceID == [sha256.Size]byte{} {
		return resourceSpec{}, nil, [sha256.Size]byte{}, ErrContract
	}
	return selected, results, resourceID, nil
}

func reportRequirement(
	aggregate bool,
	kind restauthorization.RequirementKind,
) bool {
	if !aggregate {
		switch kind {
		case restauthorization.RequirementResourceRead,
			restauthorization.RequirementFieldSelect,
			restauthorization.RequirementFieldFilter,
			restauthorization.RequirementFieldSort,
			restauthorization.RequirementExactCount:
			return true
		}
		return false
	}
	switch kind {
	case restauthorization.RequirementResourceRead,
		restauthorization.RequirementFieldFilter,
		restauthorization.RequirementFieldGroup,
		restauthorization.RequirementFieldMeasure,
		restauthorization.RequirementFieldHaving,
		restauthorization.RequirementAggregateMeasure,
		restauthorization.RequirementAggregateHaving,
		restauthorization.RequirementAggregateSort:
		return true
	default:
		return false
	}
}
