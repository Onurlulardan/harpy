package api

import "fmt"

type Format string

const (
	FormatCSV   Format = "csv"
	FormatJSONL Format = "jsonl"
	FormatXLSX  Format = "xlsx"
)

func ParseFormat(value string) (Format, error) {
	format := Format(value)
	if !format.Valid() {
		return "", fmt.Errorf("tabular transfer format is invalid")
	}
	return format, nil
}

func (format Format) Valid() bool {
	return format == FormatCSV || format == FormatJSONL || format == FormatXLSX
}

func (format Format) Importable() bool {
	return format == FormatCSV || format == FormatJSONL
}
