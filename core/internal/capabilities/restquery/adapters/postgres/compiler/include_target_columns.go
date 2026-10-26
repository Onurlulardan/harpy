package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

func includeTargetColumns(
	include resolved.Include,
) ([]recordColumn, int, error) {
	projections, projectionErr := include.Projections()
	sorts, sortErr := include.Sorts()
	stable, stableErr := include.StableID()
	stableID, stableIDErr := stable.ID()
	if projectionErr != nil || sortErr != nil || stableErr != nil || stableIDErr != nil {
		return nil, 0, ErrContract
	}
	metadata, err := recordSortMetadataByField(sorts)
	if err != nil {
		return nil, 0, err
	}
	columns := make([]recordColumn, 0, len(projections)+len(sorts))
	seen := make(map[uuid.UUID]struct{}, cap(columns))
	stableIndex := -1
	for _, projection := range projections {
		column, err := projectedColumn(projection, metadata)
		if err != nil {
			return nil, 0, err
		}
		if _, duplicate := seen[column.fieldID]; duplicate {
			return nil, 0, ErrContract
		}
		seen[column.fieldID] = struct{}{}
		if column.fieldID == stableID {
			stableIndex = len(columns)
		}
		columns = append(columns, column)
	}
	for _, sort := range sorts {
		field, err := sort.Field()
		if err != nil {
			return nil, 0, ErrContract
		}
		identifier, _, err := fieldIdentity(field)
		if err != nil {
			return nil, 0, err
		}
		if _, present := seen[identifier]; present {
			continue
		}
		column, err := hiddenIncludeSortColumn(field, metadata)
		if err != nil {
			return nil, 0, err
		}
		seen[identifier] = struct{}{}
		columns = append(columns, column)
	}
	if len(columns) < 1 || len(columns) > 56 || stableIndex < 0 {
		return nil, 0, ErrContract
	}
	return columns, stableIndex, nil
}

func hiddenIncludeSortColumn(
	field api.FieldDefinition, metadata map[uuid.UUID]recordSortMetadata,
) (recordColumn, error) {
	identifier, _, err := fieldIdentity(field)
	fieldType, typeErr := field.Type()
	value, found := metadata[identifier]
	if err != nil || typeErr != nil || !found {
		return recordColumn{}, ErrContract
	}
	return recordColumn{
		field: field, fieldID: identifier, fieldType: fieldType,
		sortOrdinal: value.ordinal, sortDirection: value.direction,
		cursorCanonicalBytes: value.canonicalBytes,
	}, nil
}
