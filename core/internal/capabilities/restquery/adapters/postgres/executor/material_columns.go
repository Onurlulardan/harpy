package executor

import (
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/rowdecode"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

func materialFieldTypes(
	columns []compiler.ResultColumn,
) ([]api.FieldType, error) {
	if len(columns) < 1 || len(columns) > 56 {
		return nil, ErrContract
	}
	result := make([]api.FieldType, len(columns))
	for index, column := range columns {
		fieldType, err := column.FieldType()
		if err != nil {
			return nil, ErrContract
		}
		result[index] = fieldType
	}
	return result, nil
}

func materialUUIDColumn(
	row []rowdecode.Value,
	stableIndex int,
) (uuid.UUID, error) {
	if stableIndex < 0 || stableIndex >= len(row) {
		return uuid.Nil, ErrContract
	}
	kind, kindErr := row[stableIndex].Kind()
	document, documentErr := row[stableIndex].ResponseJSON()
	if kindErr != nil || documentErr != nil || kind != api.LogicalTypeUUID {
		return uuid.Nil, ErrContract
	}
	var text string
	if json.Unmarshal(document, &text) != nil {
		return uuid.Nil, ErrContract
	}
	identifier, err := uuid.Parse(text)
	if err != nil || identifier == uuid.Nil || identifier.Variant() != uuid.RFC4122 {
		return uuid.Nil, ErrContract
	}
	return identifier, nil
}

func materialOrder(
	row []rowdecode.Value,
	columns []compiler.ResultColumn,
) ([][]byte, error) {
	maximum := -1
	ordinals := make([]int, len(columns))
	for index, column := range columns {
		ordinal, sorted, err := column.SortOrdinal()
		if err != nil {
			return nil, ErrContract
		}
		ordinals[index] = -1
		if sorted {
			ordinals[index] = ordinal
			maximum = max(maximum, ordinal)
		}
	}
	if maximum < 0 || maximum > 5 || len(row) != len(columns) {
		return nil, ErrContract
	}
	result := make([][]byte, maximum+1)
	for index, ordinal := range ordinals {
		if ordinal < 0 {
			continue
		}
		value, err := row[index].ResponseJSON()
		if err != nil || result[ordinal] != nil {
			return nil, ErrContract
		}
		result[ordinal] = value
	}
	for _, value := range result {
		if value == nil {
			return nil, ErrContract
		}
	}
	return result, nil
}
