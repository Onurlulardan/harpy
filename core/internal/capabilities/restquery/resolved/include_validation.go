package resolved

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

func validIncludeSorts(include Include) bool {
	cardinality, err := include.relation.Cardinality()
	if err != nil {
		return false
	}
	explicit, implicit := 0, 0
	seen := make(map[uuid.UUID]struct{}, len(include.sorts))
	for _, sort := range include.sorts {
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
		} else {
			explicit++
		}
	}
	if explicit > 5 {
		return false
	}
	if cardinality == api.RelationCardinalityToOne {
		return implicit == 0
	}
	return validStableFinalSort(include.sorts, include.stableID, implicit)
}

func validStableFinalSort(
	sorts []Sort, stableID api.FieldDefinition, implicitCount int,
) bool {
	if len(sorts) == 0 || implicitCount > 1 {
		return false
	}
	stable, valid := fieldIdentifier(stableID)
	finalID, finalValid := fieldIdentifier(sorts[len(sorts)-1].field)
	if !valid || !finalValid || stable != finalID {
		return false
	}
	final := sorts[len(sorts)-1]
	if final.implicit && final.direction != api.DirectionAscending {
		return false
	}
	for index, sort := range sorts {
		if sort.implicit && index != len(sorts)-1 {
			return false
		}
	}
	return true
}
