package recordpage

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/rowdecode"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
)

// Assemble validates exact positional rows and builds the root record page.
func Assemble(
	statement compiler.RecordStatement,
	codec cursor.Codec,
	rows [][]rowdecode.Value,
) (Page, error) {
	if codec.Validate() != nil {
		return Page{}, ErrContract
	}
	metadata, err := loadMetadata(statement)
	if err != nil || len(rows) > metadata.requestedSize+1 {
		return Page{}, ErrContract
	}
	for _, row := range rows {
		if validateRow(row, metadata.columns) != nil {
			return Page{}, ErrContract
		}
	}
	hasMore := len(rows) == metadata.requestedSize+1
	included := len(rows)
	if hasMore {
		included = metadata.requestedSize
	}
	records := make([][]byte, included)
	for index := 0; index < included; index++ {
		record, encodeErr := encodeRecord(rows[index], metadata.columns)
		if encodeErr != nil {
			return Page{}, ErrContract
		}
		records[index] = record
	}
	nextCursor := ""
	if hasMore {
		nextCursor, err = encodeNextCursor(codec, metadata, rows[included-1])
		if err != nil {
			return Page{}, ErrContract
		}
	}
	return newPage(records, hasMore, nextCursor, metadata.requestedSize)
}

func validateRow(row []rowdecode.Value, columns []columnMetadata) error {
	if len(row) != len(columns) {
		return ErrContract
	}
	for index, value := range row {
		kind, kindErr := value.Kind()
		fieldType, typeErr := value.FieldType()
		exact, exactErr := fieldType.ExactEqual(columns[index].fieldType)
		if kindErr != nil || typeErr != nil || exactErr != nil || !exact ||
			kind != columns[index].kind {
			return ErrContract
		}
	}
	return nil
}
