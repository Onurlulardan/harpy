package resolved

import (
	"slices"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func sameFieldDefinition(left, right api.FieldDefinition) bool {
	leftID, leftIDErr := left.ID()
	rightID, rightIDErr := right.ID()
	leftName, leftNameErr := left.PublicName()
	rightName, rightNameErr := right.PublicName()
	leftStorage, leftStorageErr := left.Storage()
	rightStorage, rightStorageErr := right.Storage()
	leftType, leftTypeErr := left.Type()
	rightType, rightTypeErr := right.Type()
	leftCaps, leftCapsErr := left.Capabilities()
	rightCaps, rightCapsErr := right.Capabilities()
	return leftIDErr == nil && rightIDErr == nil && leftNameErr == nil &&
		rightNameErr == nil && leftStorageErr == nil && rightStorageErr == nil &&
		leftTypeErr == nil && rightTypeErr == nil && leftCapsErr == nil &&
		rightCapsErr == nil && leftID == rightID && leftName == rightName &&
		leftStorage == rightStorage && sameFieldType(leftType, rightType) &&
		sameFieldCapabilities(leftCaps, rightCaps)
}

func sameFieldCapabilities(left, right api.FieldCapabilities) bool {
	leftLogical, leftLogicalErr := left.LogicalType()
	rightLogical, rightLogicalErr := right.LogicalType()
	leftSelectable, leftSelectableErr := left.Selectable()
	rightSelectable, rightSelectableErr := right.Selectable()
	leftFilters, leftFiltersErr := left.Filters()
	rightFilters, rightFiltersErr := right.Filters()
	leftSortable, leftSortableErr := left.Sortable()
	rightSortable, rightSortableErr := right.Sortable()
	leftBound, leftBoundErr := left.CursorCanonicalBytes()
	rightBound, rightBoundErr := right.CursorCanonicalBytes()
	leftSearchable, leftSearchableErr := left.Searchable()
	rightSearchable, rightSearchableErr := right.Searchable()
	leftGroupable, leftGroupableErr := left.Groupable()
	rightGroupable, rightGroupableErr := right.Groupable()
	leftAggregates, leftAggregatesErr := left.Aggregates()
	rightAggregates, rightAggregatesErr := right.Aggregates()
	return leftLogicalErr == nil && rightLogicalErr == nil &&
		leftSelectableErr == nil && rightSelectableErr == nil &&
		leftFiltersErr == nil && rightFiltersErr == nil &&
		leftSortableErr == nil && rightSortableErr == nil &&
		leftBoundErr == nil && rightBoundErr == nil &&
		leftSearchableErr == nil && rightSearchableErr == nil &&
		leftGroupableErr == nil && rightGroupableErr == nil &&
		leftAggregatesErr == nil && rightAggregatesErr == nil &&
		leftLogical == rightLogical && leftSelectable == rightSelectable &&
		leftSortable == rightSortable && leftBound == rightBound &&
		leftSearchable == rightSearchable && leftGroupable == rightGroupable &&
		slices.Equal(leftFilters, rightFilters) && slices.Equal(leftAggregates, rightAggregates)
}
