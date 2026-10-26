package api

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

type csvReader struct {
	parser  *csv.Reader
	columns []string
	limits  Limits
	index   int64
}

func newCSVReader(source io.Reader, limits Limits) (*csvReader, error) {
	parser := csv.NewReader(io.LimitReader(source, limits.DecompressedBytes+1))
	parser.FieldsPerRecord = -1
	header, err := parser.Read()
	if err != nil || len(header) < 1 || len(header) > limits.Columns ||
		parser.InputOffset() > limits.RowBytes {
		return nil, ErrFormat
	}
	seen := make(map[string]struct{}, len(header))
	for _, column := range header {
		if column == "" || strings.TrimSpace(column) != column {
			return nil, ErrFormat
		}
		if _, exists := seen[column]; exists {
			return nil, ErrFormat
		}
		seen[column] = struct{}{}
	}
	return &csvReader{parser: parser, columns: append([]string(nil), header...),
		limits: limits}, nil
}

func (reader *csvReader) Columns() []string {
	return append([]string(nil), reader.columns...)
}

func (reader *csvReader) Next() (Row, error) {
	start := reader.parser.InputOffset()
	values, err := reader.parser.Read()
	if err != nil {
		if err == io.EOF {
			return Row{}, io.EOF
		}
		return Row{}, ErrFormat
	}
	reader.index++
	if reader.index > int64(reader.limits.Rows) ||
		reader.parser.InputOffset()-start > reader.limits.RowBytes {
		return Row{}, ErrRowLimit
	}
	if len(values) != len(reader.columns) {
		return Row{}, fmt.Errorf("%w: CSV row width differs", ErrFormat)
	}
	row := Row{Index: reader.index, Values: make(map[string]string, len(values))}
	for index, value := range values {
		row.Values[reader.columns[index]] = value
	}
	return row, nil
}
