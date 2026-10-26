package authorization

import "fmt"

type RequirementKind uint8

const (
	RequirementResourceRead RequirementKind = iota + 1
	RequirementFieldSelect
	RequirementFieldFilter
	RequirementFieldSort
	RequirementFieldSearch
	RequirementFieldGroup
	RequirementFieldMeasure
	RequirementFieldHaving
	RequirementRelationInclude
	RequirementExactCount
	RequirementAggregateMeasure
	RequirementAggregateHaving
	RequirementAggregateSort
)

func (kind RequirementKind) String() string {
	switch kind {
	case RequirementResourceRead:
		return "resource_read"
	case RequirementFieldSelect:
		return "field_select"
	case RequirementFieldFilter:
		return "field_filter"
	case RequirementFieldSort:
		return "field_sort"
	case RequirementFieldSearch:
		return "field_search"
	case RequirementFieldGroup:
		return "field_group"
	case RequirementFieldMeasure:
		return "field_measure"
	case RequirementFieldHaving:
		return "field_having"
	case RequirementRelationInclude:
		return "relation_include"
	case RequirementExactCount:
		return "exact_count"
	case RequirementAggregateMeasure:
		return "aggregate_measure"
	case RequirementAggregateHaving:
		return "aggregate_having"
	case RequirementAggregateSort:
		return "aggregate_sort"
	default:
		return ""
	}
}

func (kind RequirementKind) Validate() error {
	if kind.String() == "" {
		return fmt.Errorf("REST Query authorization requirement kind is invalid")
	}
	return nil
}

func (RequirementKind) MarshalJSON() ([]byte, error) {
	return rejectAuthorizationJSON()
}
