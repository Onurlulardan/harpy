package systemtransferartifact

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"strings"

	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

const maximumAnalysisSampleBytes = 1 << 20

func analyzeSource(format string, captured []byte,
	truncated bool) (systemtransfer.SourceAnalysis, error) {
	captured = completeAnalysisPrefix(captured, truncated)
	var analysis systemtransfer.SourceAnalysis
	var err error
	switch format {
	case "csv":
		analysis, err = analyzeCSV(captured)
	case "jsonl":
		analysis, err = analyzeJSONL(captured)
	default:
		err = errInvalid
	}
	if err != nil || analysis.Validate() != nil {
		return systemtransfer.SourceAnalysis{}, errors.Join(errInvalid, err)
	}
	return analysis, nil
}

func analyzeCSV(captured []byte) (systemtransfer.SourceAnalysis, error) {
	reader := csv.NewReader(bytes.NewReader(captured))
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil || len(header) < 1 || len(header) > systemtransfer.MaximumSourceAnalysisColumns {
		return systemtransfer.SourceAnalysis{}, errInvalid
	}
	columns := make([]systemtransfer.SourceColumn, len(header))
	seen := make(map[string]struct{}, len(header))
	for index, raw := range header {
		name := strings.TrimSpace(raw)
		if name == "" || name != raw {
			return systemtransfer.SourceAnalysis{}, errInvalid
		}
		if _, duplicate := seen[name]; duplicate {
			return systemtransfer.SourceAnalysis{}, errInvalid
		}
		seen[name] = struct{}{}
		columns[index] = systemtransfer.SourceColumn{Name: name, Kind: "string"}
	}
	samples := make([]map[string]any, 0, systemtransfer.MaximumSourceAnalysisSamples)
	for rows := 0; rows < systemtransfer.MaximumSourceAnalysisSamples; rows++ {
		row, readErr := reader.Read()
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil || len(row) != len(header) {
			return systemtransfer.SourceAnalysis{}, errInvalid
		}
		sample := make(map[string]any, len(header))
		for index, value := range row {
			sample[header[index]] = value
		}
		samples = appendBoundedSample(samples, sample)
	}
	return systemtransfer.SourceAnalysis{Columns: columns, Samples: samples}, nil
}

func analyzeJSONL(captured []byte) (systemtransfer.SourceAnalysis, error) {
	lines := bytes.Split(captured, []byte{'\n'})
	samples := make([]map[string]any, 0, systemtransfer.MaximumSourceAnalysisSamples)
	kinds := make(map[string]string)
	rows := 0
	for _, line := range lines {
		line = bytes.TrimSuffix(line, []byte{'\r'})
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		if rows == systemtransfer.MaximumSourceAnalysisSamples {
			break
		}
		object, err := decodeJSONObject(line)
		if err != nil {
			return systemtransfer.SourceAnalysis{}, err
		}
		for name, value := range object {
			if len(name) < 1 || len(name) > 256 {
				return systemtransfer.SourceAnalysis{}, errInvalid
			}
			kinds[name] = mergeKind(kinds[name], sourceValueKind(value))
		}
		rows++
		samples = appendBoundedSample(samples, object)
	}
	if len(kinds) < 1 || len(kinds) > systemtransfer.MaximumSourceAnalysisColumns {
		return systemtransfer.SourceAnalysis{}, errInvalid
	}
	names := make([]string, 0, len(kinds))
	for name := range kinds {
		names = append(names, name)
	}
	sort.Strings(names)
	columns := make([]systemtransfer.SourceColumn, len(names))
	for index, name := range names {
		columns[index] = systemtransfer.SourceColumn{Name: name, Kind: kinds[name]}
	}
	return systemtransfer.SourceAnalysis{Columns: columns, Samples: samples}, nil
}

func completeAnalysisPrefix(value []byte, truncated bool) []byte {
	if !truncated {
		return value
	}
	lastLine := bytes.LastIndexByte(value, '\n')
	if lastLine < 0 {
		return nil
	}
	return value[:lastLine+1]
}

func appendBoundedSample(samples []map[string]any,
	sample map[string]any) []map[string]any {
	encoded, err := json.Marshal(sample)
	if err == nil && len(encoded) <= maximumAnalysisSampleBytes {
		return append(samples, sample)
	}
	return samples
}

func decodeJSONObject(value []byte) (map[string]any, error) {
	decoder := json.NewDecoder(bytes.NewReader(value))
	decoder.UseNumber()
	var object map[string]any
	if err := decoder.Decode(&object); err != nil || object == nil {
		return nil, errInvalid
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return nil, errInvalid
	}
	return object, nil
}

func sourceValueKind(value any) string {
	switch value.(type) {
	case nil:
		return "null"
	case string:
		return "string"
	case json.Number:
		return "number"
	case bool:
		return "boolean"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	default:
		return "mixed"
	}
}

func mergeKind(current, next string) string {
	if current == "" || current == next {
		return next
	}
	return "mixed"
}
