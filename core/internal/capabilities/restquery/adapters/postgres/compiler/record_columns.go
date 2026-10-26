package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

type recordColumn struct {
	field                api.FieldDefinition
	fieldID              uuid.UUID
	fieldType            api.FieldType
	outputKey            string
	hasOutput            bool
	sortOrdinal          int
	sortDirection        api.Direction
	cursorCanonicalBytes int
}

type recordSortMetadata struct {
	ordinal        int
	direction      api.Direction
	canonicalBytes int
}

func recordColumns(query resolved.RecordQuery) ([]recordColumn, error) {
	projections, projectionErr := query.Projections()
	sorts, sortErr := query.Sorts()
	if projectionErr != nil || sortErr != nil {
		return nil, ErrContract
	}
	metadata, err := recordSortMetadataByField(sorts)
	if err != nil {
		return nil, err
	}
	columns := make([]recordColumn, 0, len(projections)+len(sorts))
	seen := make(map[uuid.UUID]struct{}, cap(columns))
	for _, projection := range projections {
		column, err := projectedColumn(projection, metadata)
		if err != nil {
			return nil, err
		}
		if _, duplicate := seen[column.fieldID]; duplicate {
			return nil, ErrContract
		}
		seen[column.fieldID] = struct{}{}
		columns = append(columns, column)
	}
	for _, sort := range sorts {
		field, err := sort.Field()
		if err != nil {
			return nil, ErrContract
		}
		identifier, _, err := fieldIdentity(field)
		if err != nil {
			return nil, err
		}
		fieldType, err := field.Type()
		if err != nil {
			return nil, ErrContract
		}
		if _, present := seen[identifier]; present {
			continue
		}
		sortMetadata, present := metadata[identifier]
		if !present {
			return nil, ErrContract
		}
		seen[identifier] = struct{}{}
		columns = append(columns, recordColumn{
			field: field, fieldID: identifier, fieldType: fieldType,
			sortOrdinal:          sortMetadata.ordinal,
			sortDirection:        sortMetadata.direction,
			cursorCanonicalBytes: sortMetadata.canonicalBytes,
		})
	}
	if len(columns) < 1 || len(columns) > 56 {
		return nil, ErrContract
	}
	return columns, nil
}

func recordSortMetadataByField(
	sorts []resolved.Sort,
) (map[uuid.UUID]recordSortMetadata, error) {
	result := make(map[uuid.UUID]recordSortMetadata, len(sorts))
	for index, sort := range sorts {
		field, err := sort.Field()
		if err != nil {
			return nil, ErrContract
		}
		identifier, err := field.ID()
		if err != nil {
			return nil, ErrContract
		}
		direction, directionErr := sort.Direction()
		capabilities, capabilityErr := field.Capabilities()
		bound, boundErr := capabilities.CursorCanonicalBytes()
		if directionErr != nil || capabilityErr != nil || boundErr != nil ||
			direction.String() == "" || bound < 1 || bound > 2048 {
			return nil, ErrContract
		}
		result[identifier] = recordSortMetadata{
			ordinal: index, direction: direction, canonicalBytes: bound,
		}
	}
	return result, nil
}
