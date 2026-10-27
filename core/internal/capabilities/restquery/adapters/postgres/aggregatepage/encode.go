package aggregatepage

import (
	"bytes"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/rowdecode"
)

func encodeGroup(row []rowdecode.Value, metadata metadata) ([]byte, error) {
	if validateRow(row, metadata) != nil {
		return nil, ErrContract
	}
	var document bytes.Buffer
	document.WriteString(`{"key":{`)
	if writeMembers(&document, row[:len(metadata.groups)], metadata.groups) != nil {
		return nil, ErrContract
	}
	document.WriteString(`},"measures":{`)
	if writeMembers(&document, row[len(metadata.groups):], metadata.measures) != nil {
		return nil, ErrContract
	}
	document.WriteString("}}")
	if !json.Valid(document.Bytes()) {
		return nil, ErrContract
	}
	return document.Bytes(), nil
}

func writeMembers(
	document *bytes.Buffer, values []rowdecode.Value, columns []column,
) error {
	if document == nil || len(values) != len(columns) {
		return ErrContract
	}
	for index, column := range columns {
		key, keyErr := json.Marshal(column.key)
		value, valueErr := values[index].ResponseJSON()
		if keyErr != nil || valueErr != nil || !json.Valid(value) {
			return ErrContract
		}
		if index > 0 {
			document.WriteByte(',')
		}
		document.Write(key)
		document.WriteByte(':')
		document.Write(value)
	}
	return nil
}
