package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/recordpage"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/rowdecode"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func readRecords(
	ctx context.Context,
	driver execution.QueryDriver,
	policy planner.Policy,
	compiled compiler.RecordStatement,
	statement statement,
	codec cursor.Codec,
) (recordpage.Page, error) {
	fieldTypes, requested, err := recordShape(compiled)
	if err != nil {
		return recordpage.Page{}, err
	}
	if err := applyTimeout(ctx, driver, policy, statement.class); err != nil {
		return recordpage.Page{}, err
	}
	rows, err := driver.Query(ctx, statement.sql, statement.arguments...)
	if err != nil {
		return recordpage.Page{}, mapDatabaseError(ctx, policy, statement.class, err)
	}
	if interfaceIsNil(rows) {
		return recordpage.Page{}, ErrContract
	}
	defer rows.Close()
	decoded := make([][]rowdecode.Value, 0, requested+1)
	for rows.Next() {
		if len(decoded) == requested+1 {
			return recordpage.Page{}, ErrContract
		}
		row, err := scanRecordRow(rows, fieldTypes)
		if err != nil {
			return recordpage.Page{}, mapScanError(ctx, policy, statement.class, err)
		}
		decoded = append(decoded, row)
	}
	if err := rows.Err(); err != nil {
		return recordpage.Page{}, mapDatabaseError(ctx, policy, statement.class, err)
	}
	page, err := recordpage.Assemble(compiled, codec, decoded)
	if err != nil {
		return recordpage.Page{}, ErrContract
	}
	return page, nil
}

func recordShape(statement compiler.RecordStatement) ([]api.FieldType, int, error) {
	columns, columnsErr := statement.Columns()
	requested, requestedErr := statement.RequestedSize()
	if columnsErr != nil || requestedErr != nil {
		return nil, 0, ErrContract
	}
	fieldTypes := make([]api.FieldType, len(columns))
	for index, column := range columns {
		fieldType, err := column.FieldType()
		if err != nil {
			return nil, 0, ErrContract
		}
		fieldTypes[index] = fieldType
	}
	return fieldTypes, requested, nil
}

func scanRecordRow(
	rows execution.Rows,
	fieldTypes []api.FieldType,
) ([]rowdecode.Value, error) {
	scanner, err := rowdecode.NewRowScanner(fieldTypes)
	if err != nil {
		return nil, ErrContract
	}
	destinations, err := scanner.Destinations()
	if err != nil {
		return nil, ErrContract
	}
	if err := rows.Scan(destinations...); err != nil {
		return nil, err
	}
	row, err := scanner.Decode()
	if err != nil {
		return nil, ErrContract
	}
	return row, nil
}
