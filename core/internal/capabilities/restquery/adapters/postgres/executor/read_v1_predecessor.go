package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/rowdecode"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func readV1Predecessor(
	ctx context.Context,
	driver execution.SnapshotDriver,
	policy planner.Policy,
	compiled compiler.PredecessorStatement,
	candidate statement,
) (compiler.PredecessorTuple, bool, error) {
	fieldTypes, err := predecessorFieldTypes(compiled)
	if err != nil || candidate.class != planner.StatementClassV1PredecessorSeek {
		return compiler.PredecessorTuple{}, false, ErrContract
	}
	if err := applyTimeout(ctx, driver, policy, candidate.class); err != nil {
		return compiler.PredecessorTuple{}, false, err
	}
	rows, err := driver.Query(ctx, candidate.sql, candidate.arguments...)
	if err != nil {
		return compiler.PredecessorTuple{}, false,
			mapDatabaseError(ctx, policy, candidate.class, err)
	}
	if interfaceIsNil(rows) {
		return compiler.PredecessorTuple{}, false, ErrContract
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return compiler.PredecessorTuple{}, false,
				mapDatabaseError(ctx, policy, candidate.class, err)
		}
		return compiler.PredecessorTuple{}, false, nil
	}
	row, err := scanRecordRow(rows, fieldTypes)
	if err != nil {
		return compiler.PredecessorTuple{}, false,
			mapScanError(ctx, policy, candidate.class, err)
	}
	if rows.Next() {
		return compiler.PredecessorTuple{}, false, ErrContract
	}
	if err := rows.Err(); err != nil {
		return compiler.PredecessorTuple{}, false,
			mapDatabaseError(ctx, policy, candidate.class, err)
	}
	values, err := predecessorLastValues(row)
	if err != nil {
		return compiler.PredecessorTuple{}, false, ErrContract
	}
	tuple, err := compiled.MintTuple(values)
	if err != nil {
		return compiler.PredecessorTuple{}, false, ErrContract
	}
	return tuple, true, nil
}

func predecessorFieldTypes(
	statement compiler.PredecessorStatement,
) ([]api.FieldType, error) {
	columns, err := statement.Columns()
	if err != nil {
		return nil, ErrContract
	}
	types := make([]api.FieldType, len(columns))
	for index, column := range columns {
		types[index], err = column.FieldType()
		if err != nil {
			return nil, ErrContract
		}
	}
	return types, nil
}

func predecessorLastValues(row []rowdecode.Value) ([]cursor.LastValue, error) {
	values := make([]cursor.LastValue, len(row))
	for index, value := range row {
		last, present, err := value.LastValue()
		if err != nil || !present {
			return nil, ErrContract
		}
		values[index] = last
	}
	return values, nil
}
