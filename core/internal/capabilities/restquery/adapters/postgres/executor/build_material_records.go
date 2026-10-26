package executor

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/rowdecode"
)

func buildMaterialRecords(
	columns []compiler.ResultColumn,
	stableIndex int,
	rows [][]rowdecode.Value,
	documents [][]byte,
) ([]materialRecord, error) {
	if len(rows) != len(documents) || rows == nil || documents == nil ||
		stableIndex < 0 || stableIndex >= len(columns) {
		return nil, ErrContract
	}
	result := make([]materialRecord, len(rows))
	seen := make(map[[16]byte]struct{}, len(rows))
	for index, row := range rows {
		identifier, err := materialUUIDColumn(row, stableIndex)
		if err != nil {
			return nil, ErrContract
		}
		if _, duplicate := seen[identifier]; duplicate {
			return nil, ErrContract
		}
		seen[identifier] = struct{}{}
		order, err := materialOrder(row, columns)
		if err != nil {
			return nil, ErrContract
		}
		result[index], err = newMaterialRecord(identifier, documents[index], order)
		if err != nil {
			return nil, ErrContract
		}
	}
	return result, nil
}
