package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/recordpage"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func readGraphRoot(
	ctx context.Context,
	driver execution.SnapshotDriver,
	policy planner.Policy,
	compiled compiler.RecordStatement,
	candidate statement,
	codec cursor.Codec,
	stableIndex int,
) (recordpage.Page, []materialRecord, error) {
	columns, columnErr := compiled.Columns()
	requested, requestedErr := compiled.RequestedSize()
	fieldTypes, typeErr := materialFieldTypes(columns)
	if columnErr != nil || requestedErr != nil || typeErr != nil ||
		stableIndex < 0 || stableIndex >= len(columns) {
		return recordpage.Page{}, nil, ErrContract
	}
	rows, err := readMaterialRows(
		ctx, driver, policy, candidate, fieldTypes, requested+1,
	)
	if err != nil {
		return recordpage.Page{}, nil, err
	}
	page, err := recordpage.Assemble(compiled, codec, rows)
	if err != nil {
		return recordpage.Page{}, nil, ErrContract
	}
	documents, err := page.Records()
	if err != nil || len(documents) > len(rows) {
		return recordpage.Page{}, nil, ErrContract
	}
	records, err := buildMaterialRecords(
		columns, stableIndex, rows[:len(documents)], documents,
	)
	if err != nil {
		return recordpage.Page{}, nil, err
	}
	return page, records, nil
}
