package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

func projectedColumn(
	projection resolved.Projection, metadata map[uuid.UUID]recordSortMetadata,
) (recordColumn, error) {
	field, fieldErr := projection.Field()
	key, keyErr := projection.Key()
	keyValue, keyValueErr := key.Value()
	hidden, hiddenErr := projection.Hidden()
	identifier, _, identityErr := fieldIdentity(field)
	fieldType, typeErr := field.Type()
	if fieldErr != nil || keyErr != nil || keyValueErr != nil || hiddenErr != nil ||
		identityErr != nil || typeErr != nil {
		return recordColumn{}, ErrContract
	}
	sortMetadata := recordSortMetadata{ordinal: -1}
	if value, found := metadata[identifier]; found {
		sortMetadata = value
	}
	outputKey := ""
	if !hidden {
		outputKey = keyValue
	}
	return recordColumn{
		field: field, fieldID: identifier, fieldType: fieldType,
		outputKey: outputKey, hasOutput: !hidden,
		sortOrdinal: sortMetadata.ordinal, sortDirection: sortMetadata.direction,
		cursorCanonicalBytes: sortMetadata.canonicalBytes,
	}, nil
}

func resultColumns(columns []recordColumn) []ResultColumn {
	result := make([]ResultColumn, len(columns))
	for index, column := range columns {
		result[index] = ResultColumn{
			fieldID: column.fieldID, fieldType: column.fieldType,
			outputKey: column.outputKey, hasOutput: column.hasOutput,
			sortOrdinal: column.sortOrdinal, sortDirection: column.sortDirection,
			cursorCanonicalBytes: column.cursorCanonicalBytes, validated: true,
		}
	}
	return result
}
