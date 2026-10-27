package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

func v1PredecessorColumns(query resolved.RecordQuery) ([]recordColumn, error) {
	sorts, err := query.Sorts()
	if err != nil || len(sorts) == 0 || len(sorts) > 6 {
		return nil, ErrContract
	}
	metadata, err := recordSortMetadataByField(sorts)
	if err != nil {
		return nil, err
	}
	columns := make([]recordColumn, len(sorts))
	seen := make(map[uuid.UUID]struct{}, len(sorts))
	for index, sort := range sorts {
		field, fieldErr := sort.Field()
		identifier, _, identityErr := fieldIdentity(field)
		fieldType, typeErr := field.Type()
		value, found := metadata[identifier]
		if fieldErr != nil || identityErr != nil || typeErr != nil || !found ||
			value.ordinal != index {
			return nil, ErrContract
		}
		if _, duplicate := seen[identifier]; duplicate {
			return nil, ErrContract
		}
		seen[identifier] = struct{}{}
		columns[index] = recordColumn{
			field: field, fieldID: identifier, fieldType: fieldType,
			sortOrdinal: index, sortDirection: value.direction,
			cursorCanonicalBytes: value.canonicalBytes,
		}
	}
	return columns, nil
}
