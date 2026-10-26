package recordpage

import (
	"bytes"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/rowdecode"
)

func encodeRecord(
	row []rowdecode.Value,
	columns []columnMetadata,
) ([]byte, error) {
	if validateRow(row, columns) != nil {
		return nil, ErrContract
	}
	var document bytes.Buffer
	document.WriteByte('{')
	written := 0
	for index, column := range columns {
		if !column.hasOutput {
			continue
		}
		key, err := json.Marshal(column.outputKey)
		if err != nil {
			return nil, ErrContract
		}
		value, err := row[index].ResponseJSON()
		if err != nil || !json.Valid(value) {
			return nil, ErrContract
		}
		if written > 0 {
			document.WriteByte(',')
		}
		document.Write(key)
		document.WriteByte(':')
		document.Write(value)
		written++
	}
	document.WriteByte('}')
	return document.Bytes(), nil
}
