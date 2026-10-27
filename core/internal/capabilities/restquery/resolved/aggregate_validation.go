package resolved

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func validAggregateHaving(
	having *Having, groups []AggregateGroup, measures []AggregateMeasure,
) bool {
	if having == nil {
		return true
	}
	if having.Validate() != nil {
		return false
	}
	return havingReferencesValid(*having, groups, measures)
}

func havingReferencesValid(
	having Having, groups []AggregateGroup, measures []AggregateMeasure,
) bool {
	if having.predicate != nil {
		fieldType, found := aggregateReferenceType(
			having.predicate.reference, groups, measures,
		)
		if !found {
			return false
		}
		if !operatorAllowedForType(fieldType, having.predicate.operator) {
			return false
		}
		for _, value := range having.predicate.values {
			if !valueConformsToFieldType(
				fieldType, having.predicate.operator, value,
			) {
				return false
			}
		}
		return true
	}
	for _, child := range having.children {
		if !havingReferencesValid(child, groups, measures) {
			return false
		}
	}
	return true
}

func validAggregateSorts(sorts []AggregateSort, groupCount, measureCount int) bool {
	type referenceKey struct {
		kind    api.AggregateReference
		ordinal int
	}
	seen := make(map[referenceKey]struct{}, len(sorts))
	explicitGroups := make(map[int]struct{}, groupCount)
	implicitStart := len(sorts)
	explicitCount := 0
	for index, sort := range sorts {
		if sort.Validate() != nil ||
			!referenceInBounds(sort.reference, groupCount, measureCount) {
			return false
		}
		key := referenceKey{kind: sort.reference.kind, ordinal: sort.reference.ordinal}
		if _, duplicate := seen[key]; duplicate {
			return false
		}
		seen[key] = struct{}{}
		if sort.implicit {
			implicitStart = min(implicitStart, index)
			continue
		}
		if implicitStart != len(sorts) {
			return false
		}
		explicitCount++
		if sort.reference.kind == api.AggregateReferenceGroup {
			explicitGroups[sort.reference.ordinal] = struct{}{}
		}
	}
	if explicitCount > 3 {
		return false
	}
	expected := make([]int, 0, groupCount-len(explicitGroups))
	for ordinal := 0; ordinal < groupCount; ordinal++ {
		if _, explicit := explicitGroups[ordinal]; !explicit {
			expected = append(expected, ordinal)
		}
	}
	implicit := sorts[implicitStart:]
	if len(implicit) != len(expected) {
		return false
	}
	for index, sort := range implicit {
		if sort.reference.kind != api.AggregateReferenceGroup ||
			sort.reference.ordinal != expected[index] ||
			sort.direction != api.DirectionAscending {
			return false
		}
	}
	return true
}

func referenceInBounds(reference AggregateReference, groupCount, measureCount int) bool {
	if reference.kind == api.AggregateReferenceGroup {
		return reference.ordinal < groupCount
	}
	return reference.kind == api.AggregateReferenceMeasure &&
		reference.ordinal < measureCount
}

func aggregateReferenceType(
	reference AggregateReference,
	groups []AggregateGroup,
	measures []AggregateMeasure,
) (api.FieldType, bool) {
	if reference.Validate() != nil ||
		!referenceInBounds(reference, len(groups), len(measures)) {
		return api.FieldType{}, false
	}
	if reference.kind == api.AggregateReferenceGroup {
		fieldType, err := groups[reference.ordinal].field.Type()
		return fieldType, err == nil
	}
	fieldType, err := measures[reference.ordinal].resultType()
	return fieldType, err == nil
}
