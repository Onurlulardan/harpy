package api

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"
)

type sliceSource struct {
	rows [][]Cell
	next int
}

func (source *sliceSource) Next() ([]Cell, error) {
	if source.next == len(source.rows) {
		return nil, io.EOF
	}
	row := source.rows[source.next]
	source.next++
	return row, nil
}

func TestWriteSupportsClosedFormats(t *testing.T) {
	for _, format := range []Format{FormatCSV, FormatJSONL, FormatXLSX} {
		var output bytes.Buffer
		result, err := Write(&output, format, []string{"id", "active"},
			&sliceSource{rows: [][]Cell{{{Kind: CellText, Text: "a"},
				{Kind: CellBoolean, Text: "true"}}}})
		if err != nil || result.Rows != 1 || output.Len() == 0 {
			t.Fatalf("format=%s result=%+v err=%v", format, result, err)
		}
		if format == FormatXLSX {
			if _, err = zip.NewReader(bytes.NewReader(output.Bytes()), int64(output.Len())); err != nil {
				t.Fatalf("xlsx archive: %v", err)
			}
		}
	}
}
