package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"unicode/utf8"
)

type CellKind string

const (
	CellNull    CellKind = "null"
	CellText    CellKind = "text"
	CellNumber  CellKind = "number"
	CellBoolean CellKind = "boolean"
	CellJSON    CellKind = "json"
)

type Cell struct {
	Kind CellKind
	Text string
}

func (cell Cell) Validate() error {
	if !utf8.ValidString(cell.Text) {
		return fmt.Errorf("tabular transfer cell is not UTF-8")
	}
	switch cell.Kind {
	case CellNull:
		if cell.Text != "" {
			return fmt.Errorf("tabular transfer null cell has content")
		}
	case CellText:
	case CellNumber:
		decoder := json.NewDecoder(bytes.NewBufferString(cell.Text))
		decoder.UseNumber()
		var number json.Number
		if err := decoder.Decode(&number); err != nil {
			return fmt.Errorf("tabular transfer number is invalid")
		}
		if _, err := number.Float64(); err != nil {
			return fmt.Errorf("tabular transfer number is invalid")
		}
	case CellBoolean:
		if cell.Text != "true" && cell.Text != "false" {
			return fmt.Errorf("tabular transfer boolean is invalid")
		}
	case CellJSON:
		if !json.Valid([]byte(cell.Text)) {
			return fmt.Errorf("tabular transfer JSON is invalid")
		}
	default:
		return fmt.Errorf("tabular transfer cell kind is invalid")
	}
	return nil
}

type RowSource interface {
	Next() ([]Cell, error)
}

type WriteResult struct {
	Rows         int64
	EscapedCells int64
}
