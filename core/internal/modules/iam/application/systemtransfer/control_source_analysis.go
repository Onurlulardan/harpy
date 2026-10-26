package systemtransfer

import (
	"encoding/json"
)

const (
	MaximumSourceAnalysisColumns = 500
	MaximumSourceAnalysisSamples = 5
	maximumSourceSampleBytes     = 1 << 20
)

type SourceColumn struct {
	Name string
	Kind string
}

type SourceAnalysis struct {
	Columns []SourceColumn
	Samples []map[string]any
}

func (analysis SourceAnalysis) Validate() error {
	if len(analysis.Columns) < 1 || len(analysis.Columns) > MaximumSourceAnalysisColumns ||
		len(analysis.Samples) > MaximumSourceAnalysisSamples {
		return ErrControlInvalid
	}
	names := make(map[string]struct{}, len(analysis.Columns))
	for _, column := range analysis.Columns {
		if len(column.Name) < 1 || len(column.Name) > 256 || !sourceKindValid(column.Kind) {
			return ErrControlInvalid
		}
		if _, duplicate := names[column.Name]; duplicate {
			return ErrControlInvalid
		}
		names[column.Name] = struct{}{}
	}
	for _, sample := range analysis.Samples {
		if len(sample) > len(names) {
			return ErrControlInvalid
		}
		for name := range sample {
			if _, known := names[name]; !known {
				return ErrControlInvalid
			}
		}
		encoded, err := json.Marshal(sample)
		if err != nil || len(encoded) > maximumSourceSampleBytes {
			return ErrControlInvalid
		}
	}
	return nil
}

func (analysis SourceAnalysis) OwnedCopy() SourceAnalysis {
	owned := SourceAnalysis{Columns: append([]SourceColumn(nil), analysis.Columns...),
		Samples: make([]map[string]any, len(analysis.Samples))}
	for index, sample := range analysis.Samples {
		encoded, err := json.Marshal(sample)
		if err != nil || json.Unmarshal(encoded, &owned.Samples[index]) != nil {
			return SourceAnalysis{}
		}
	}
	return owned
}

func sourceKindValid(value string) bool {
	return value == "null" || value == "string" || value == "number" ||
		value == "boolean" || value == "array" || value == "object" || value == "mixed"
}
