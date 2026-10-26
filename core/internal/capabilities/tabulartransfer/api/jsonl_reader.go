package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"sort"
)

type jsonlReader struct {
	scanner *bufio.Scanner
	limits  Limits
	columns []string
	index   int64
}

func newJSONLReader(source io.Reader, limits Limits) *jsonlReader {
	scanner := bufio.NewScanner(io.LimitReader(source, limits.DecompressedBytes+1))
	scanner.Buffer(make([]byte, 64*1024), int(limits.RowBytes)+1)
	return &jsonlReader{scanner: scanner, limits: limits}
}

func (reader *jsonlReader) Columns() []string {
	return append([]string(nil), reader.columns...)
}

func (reader *jsonlReader) Next() (Row, error) {
	if !reader.scanner.Scan() {
		if reader.scanner.Err() != nil {
			return Row{}, ErrRowLimit
		}
		return Row{}, io.EOF
	}
	reader.index++
	if reader.index > int64(reader.limits.Rows) ||
		int64(len(reader.scanner.Bytes())) > reader.limits.RowBytes {
		return Row{}, ErrRowLimit
	}
	decoder := json.NewDecoder(bytes.NewReader(reader.scanner.Bytes()))
	decoder.UseNumber()
	var document map[string]json.RawMessage
	if err := decoder.Decode(&document); err != nil || len(document) < 1 ||
		len(document) > reader.limits.Columns {
		return Row{}, ErrFormat
	}
	keys := make([]string, 0, len(document))
	row := Row{Index: reader.index, Values: make(map[string]string, len(document))}
	for key, raw := range document {
		if key == "" || !json.Valid(raw) {
			return Row{}, ErrFormat
		}
		keys = append(keys, key)
		row.Values[key] = string(raw)
	}
	sort.Strings(keys)
	if reader.columns == nil {
		reader.columns = keys
	} else if !sameColumns(reader.columns, keys) {
		return Row{}, ErrFormat
	}
	return row, nil
}

func sameColumns(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
