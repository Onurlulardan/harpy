package api

import (
	"fmt"
	"io"
)

func Write(destination io.Writer, format Format, columns []string,
	source RowSource) (WriteResult, error) {
	if destination == nil || source == nil || !format.Valid() ||
		len(columns) < 1 || len(columns) > MaximumColumns {
		return WriteResult{}, fmt.Errorf("tabular transfer writer dependencies are invalid")
	}
	if err := validateColumns(columns); err != nil {
		return WriteResult{}, err
	}
	switch format {
	case FormatCSV:
		return writeCSV(destination, columns, source)
	case FormatJSONL:
		return writeJSONL(destination, columns, source)
	case FormatXLSX:
		return writeXLSX(destination, columns, source)
	default:
		return WriteResult{}, ErrFormat
	}
}

func validateColumns(columns []string) error {
	seen := make(map[string]struct{}, len(columns))
	for _, column := range columns {
		if column == "" {
			return fmt.Errorf("tabular transfer column is empty")
		}
		if _, exists := seen[column]; exists {
			return fmt.Errorf("tabular transfer column is duplicated")
		}
		seen[column] = struct{}{}
	}
	return nil
}

func validateCells(columns []string, cells []Cell) error {
	if len(cells) != len(columns) {
		return fmt.Errorf("tabular transfer row width differs")
	}
	for _, cell := range cells {
		if err := cell.Validate(); err != nil {
			return err
		}
	}
	return nil
}
