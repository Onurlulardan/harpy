package recordpage

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/google/uuid"
)

func validateMetadata(
	metadata statementMetadata,
	sorts []cursor.SortDefinition,
) error {
	if metadata.requestedSize < 1 || metadata.requestedSize > 500 ||
		len(metadata.columns) < 1 || len(metadata.columns) > 56 ||
		len(metadata.sorts) != len(sorts) || len(sorts) < 1 || len(sorts) > 6 {
		return ErrContract
	}
	seenKeys := make(map[string]struct{}, len(metadata.columns))
	seenSorts := make([]bool, len(sorts))
	for index, sort := range sorts {
		if !exactSortDefinition(metadata.sorts[index], sort) {
			return ErrContract
		}
	}
	for _, column := range metadata.columns {
		declaredKind, typeErr := column.fieldType.Kind()
		if column.fieldID == uuid.Nil || column.fieldID.Variant() != uuid.RFC4122 ||
			column.fieldType.Validate() != nil ||
			typeErr != nil || declaredKind != column.kind || !column.kind.Valid() ||
			column.hasOutput != (column.outputKey != "") {
			return ErrContract
		}
		if column.hasOutput {
			if !validOutputKey(column.outputKey) {
				return ErrContract
			}
			if _, duplicate := seenKeys[column.outputKey]; duplicate {
				return ErrContract
			}
			seenKeys[column.outputKey] = struct{}{}
		}
		if !column.sorted {
			if column.sortOrdinal != -1 {
				return ErrContract
			}
			continue
		}
		if column.sortOrdinal < 0 || column.sortOrdinal >= len(sorts) ||
			seenSorts[column.sortOrdinal] ||
			!matchesSort(column, metadata.sorts[column.sortOrdinal]) {
			return ErrContract
		}
		seenSorts[column.sortOrdinal] = true
	}
	for _, seen := range seenSorts {
		if !seen {
			return ErrContract
		}
	}
	if metadata.sorts[len(metadata.sorts)-1].kind != api.LogicalTypeUUID {
		return ErrContract
	}
	return nil
}

func exactSortDefinition(expected sortMetadata, sort cursor.SortDefinition) bool {
	fieldID, idErr := sort.FieldID()
	direction, directionErr := sort.Direction()
	kind, typeErr := sort.LogicalType()
	bound, boundErr := sort.CanonicalBound()
	if idErr != nil || directionErr != nil || typeErr != nil || boundErr != nil {
		return false
	}
	return expected.fieldID == fieldID && expected.direction == direction &&
		expected.kind == kind && expected.canonicalBound == bound
}

func matchesSort(column columnMetadata, sort sortMetadata) bool {
	return sort.fieldID == column.fieldID && sort.kind == column.kind
}
