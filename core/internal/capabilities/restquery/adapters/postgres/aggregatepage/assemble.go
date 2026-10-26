package aggregatepage

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/rowdecode"
)

func Assemble(
	statement compiler.AggregateStatement,
	rows [][]rowdecode.Value,
) (Page, error) {
	metadata, err := loadMetadata(statement)
	if err != nil || rows == nil || len(rows) > metadata.limit {
		return Page{}, ErrContract
	}
	groups := make([][]byte, len(rows))
	for index, row := range rows {
		if validateRow(row, metadata) != nil {
			return Page{}, ErrContract
		}
		encoded, err := encodeGroup(row, metadata)
		if err != nil {
			return Page{}, ErrContract
		}
		groups[index] = encoded
	}
	return newPage(groups, metadata.limit)
}

func validateRow(row []rowdecode.Value, metadata metadata) error {
	columns := append(append([]column{}, metadata.groups...), metadata.measures...)
	if len(row) != len(columns) {
		return ErrContract
	}
	for index, value := range row {
		fieldType, typeErr := value.FieldType()
		exact, exactErr := fieldType.ExactEqual(columns[index].fieldType)
		if typeErr != nil || exactErr != nil || !exact {
			return ErrContract
		}
		if index >= len(metadata.groups) {
			response, err := value.ResponseJSON()
			if err != nil || !measureJSON(response) {
				return ErrContract
			}
		}
	}
	return nil
}

func measureJSON(value []byte) bool {
	return string(value) == "null" || len(value) >= 2 && value[0] == '"' &&
		value[len(value)-1] == '"'
}
