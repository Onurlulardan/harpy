package resolved

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

func validRecordSorts(sorts []Sort) bool {
	if len(sorts) < 1 || len(sorts) > 6 {
		return false
	}
	explicit, implicit := 0, 0
	seen := make(map[uuid.UUID]struct{}, len(sorts))
	for index, sort := range sorts {
		if sort.Validate() != nil {
			return false
		}
		identifier, valid := fieldIdentifier(sort.field)
		if !valid {
			return false
		}
		if _, duplicate := seen[identifier]; duplicate {
			return false
		}
		seen[identifier] = struct{}{}
		if sort.implicit {
			implicit++
			if index != len(sorts)-1 || sort.direction != api.DirectionAscending {
				return false
			}
		} else {
			explicit++
		}
	}
	return explicit <= 5 && implicit <= 1
}

func validRecordSearch(terms []SearchTerm) bool {
	if len(terms) > 16 {
		return false
	}
	seen := make(map[uuid.UUID]struct{}, len(terms))
	for _, term := range terms {
		if term.Validate() != nil {
			return false
		}
		identifier, valid := fieldIdentifier(term.field)
		if !valid {
			return false
		}
		if _, duplicate := seen[identifier]; duplicate {
			return false
		}
		seen[identifier] = struct{}{}
	}
	return true
}

func validExactCount(query RecordQuery) bool {
	if query.count != api.CountModeExact {
		return true
	}
	if !query.exactCountEnabled {
		return false
	}
	if !countSafeSearch(query.search) {
		return false
	}
	if query.where == nil {
		return true
	}
	metrics := measureFilter(*query.where, 0)
	return metrics.valid && metrics.depth <= 3 && metrics.predicates <= 20 &&
		filterCountSafe(*query.where)
}

func countSafeSearch(terms []SearchTerm) bool {
	for _, term := range terms {
		capabilities, err := term.field.Capabilities()
		if err != nil {
			return false
		}
		capability, found, err := capabilities.Filter(api.OperatorContains)
		if err != nil || !found {
			return false
		}
		countSafe, err := capability.CountSafe()
		if err != nil || !countSafe {
			return false
		}
	}
	return true
}
