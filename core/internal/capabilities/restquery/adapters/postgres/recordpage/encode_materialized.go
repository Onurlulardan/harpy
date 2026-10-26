package recordpage

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/rowdecode"
	"github.com/google/uuid"
)

// EncodeMaterialized encodes visible columns while retaining no hidden values.
func EncodeMaterialized(
	columns []compiler.ResultColumn,
	rows [][]rowdecode.Value,
) ([][]byte, error) {
	metadata, err := materializedColumns(columns)
	if err != nil || rows == nil {
		return nil, ErrContract
	}
	result := make([][]byte, len(rows))
	for index, row := range rows {
		document, err := encodeRecord(row, metadata)
		if err != nil {
			return nil, ErrContract
		}
		result[index] = document
	}
	return result, nil
}

func materializedColumns(
	columns []compiler.ResultColumn,
) ([]columnMetadata, error) {
	if len(columns) < 1 || len(columns) > 56 {
		return nil, ErrContract
	}
	result := make([]columnMetadata, len(columns))
	fields := make(map[uuid.UUID]struct{}, len(columns))
	keys := make(map[string]struct{}, len(columns))
	visible := 0
	for index, column := range columns {
		loaded, err := loadColumn(column)
		if err != nil || loaded.fieldID == uuid.Nil ||
			loaded.fieldType.Validate() != nil || loaded.kind.String() == "" {
			return nil, ErrContract
		}
		if _, duplicate := fields[loaded.fieldID]; duplicate {
			return nil, ErrContract
		}
		fields[loaded.fieldID] = struct{}{}
		if loaded.hasOutput {
			if !validOutputKey(loaded.outputKey) {
				return nil, ErrContract
			}
			if _, duplicate := keys[loaded.outputKey]; duplicate {
				return nil, ErrContract
			}
			keys[loaded.outputKey] = struct{}{}
			visible++
		}
		result[index] = loaded
	}
	if visible < 1 || visible > 50 {
		return nil, ErrContract
	}
	return result, nil
}
