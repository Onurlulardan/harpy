package recordpage

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/rowdecode"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
)

func encodeNextCursor(
	codec cursor.Codec,
	metadata statementMetadata,
	row []rowdecode.Value,
) (string, error) {
	if validateRow(row, metadata.columns) != nil || len(metadata.sorts) < 1 {
		return "", ErrContract
	}
	last := make([]cursor.LastValue, len(metadata.sorts))
	seen := make([]bool, len(metadata.sorts))
	for index, column := range metadata.columns {
		if !column.sorted {
			continue
		}
		value, present, err := row[index].LastValue()
		if err != nil || !present || value.Validate() != nil || seen[column.sortOrdinal] {
			return "", ErrContract
		}
		last[column.sortOrdinal] = value
		seen[column.sortOrdinal] = true
	}
	for _, present := range seen {
		if !present {
			return "", ErrContract
		}
	}
	stableNull, err := last[len(last)-1].IsNull()
	if err != nil || stableNull {
		return "", ErrContract
	}
	token, err := codec.Encode(metadata.context, last)
	if err != nil || token == "" {
		return "", ErrContract
	}
	return token, nil
}
