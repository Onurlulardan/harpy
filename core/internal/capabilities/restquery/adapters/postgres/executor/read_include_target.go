package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/recordpage"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func readIncludeTarget(
	ctx context.Context,
	driver execution.SnapshotDriver,
	policy planner.Policy,
	compiled compiler.IncludeTargetStatement,
	candidate statement,
	maximum int,
) ([]materialRecord, error) {
	columns, columnErr := compiled.Columns()
	stableIndex, stableErr := compiled.StableIDColumnIndex()
	fieldTypes, typeErr := materialFieldTypes(columns)
	if columnErr != nil || stableErr != nil || typeErr != nil || maximum < 1 {
		return nil, ErrContract
	}
	rows, err := readMaterialRows(
		ctx, driver, policy, candidate, fieldTypes, maximum,
	)
	if err != nil {
		return nil, err
	}
	documents, err := recordpage.EncodeMaterialized(columns, rows)
	if err != nil {
		return nil, ErrContract
	}
	return buildMaterialRecords(columns, stableIndex, rows, documents)
}
