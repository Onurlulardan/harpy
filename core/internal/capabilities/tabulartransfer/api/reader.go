package api

import (
	"fmt"
	"io"
)

func NewReader(format Format, source io.Reader, limits Limits) (Reader, error) {
	if source == nil || !format.Importable() || limits.Validate() != nil {
		return nil, fmt.Errorf("tabular transfer reader dependencies are invalid")
	}
	if format == FormatCSV {
		return newCSVReader(source, limits)
	}
	return newJSONLReader(source, limits), nil
}

func Analyze(format Format, source io.Reader, limits Limits) (Analysis, error) {
	reader, err := NewReader(format, source, limits)
	if err != nil {
		return Analysis{}, err
	}
	result := Analysis{Columns: reader.Columns()}
	for len(result.Samples) < limits.AnalysisRows {
		row, readErr := reader.Next()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return Analysis{}, readErr
		}
		result.Samples = append(result.Samples, row.Clone())
		if len(result.Columns) == 0 {
			result.Columns = reader.Columns()
		}
	}
	return result, nil
}
