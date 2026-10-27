package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/aggregatepage"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/rowdecode"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func readAggregate(
	ctx context.Context,
	driver execution.QueryDriver,
	policy planner.Policy,
	compiled compiler.AggregateStatement,
	statement statement,
) (aggregatepage.Page, error) {
	fieldTypes, limit, err := aggregateShape(compiled)
	if err != nil {
		return aggregatepage.Page{}, err
	}
	if err := applyTimeout(ctx, driver, policy, statement.class); err != nil {
		return aggregatepage.Page{}, err
	}
	rows, err := driver.Query(ctx, statement.sql, statement.arguments...)
	if err != nil {
		return aggregatepage.Page{}, mapDatabaseError(ctx, policy, statement.class, err)
	}
	if interfaceIsNil(rows) {
		return aggregatepage.Page{}, ErrContract
	}
	defer rows.Close()
	decoded := make([][]rowdecode.Value, 0, limit)
	for rows.Next() {
		if len(decoded) == limit {
			return aggregatepage.Page{}, ErrContract
		}
		row, err := scanAggregateRow(rows, fieldTypes)
		if err != nil {
			return aggregatepage.Page{}, mapScanError(ctx, policy, statement.class, err)
		}
		decoded = append(decoded, row)
	}
	if err := rows.Err(); err != nil {
		return aggregatepage.Page{}, mapDatabaseError(ctx, policy, statement.class, err)
	}
	page, err := aggregatepage.Assemble(compiled, decoded)
	if err != nil {
		return aggregatepage.Page{}, ErrContract
	}
	return page, nil
}

func aggregateShape(
	statement compiler.AggregateStatement,
) ([]api.FieldType, int, error) {
	groups, groupErr := statement.Groups()
	measures, measureErr := statement.Measures()
	limit, limitErr := statement.Limit()
	if groupErr != nil || measureErr != nil || limitErr != nil {
		return nil, 0, ErrContract
	}
	fieldTypes := make([]api.FieldType, 0, len(groups)+len(measures))
	for _, group := range groups {
		fieldType, err := group.FieldType()
		if err != nil {
			return nil, 0, ErrContract
		}
		fieldTypes = append(fieldTypes, fieldType)
	}
	for _, measure := range measures {
		fieldType, err := measure.FieldType()
		if err != nil {
			return nil, 0, ErrContract
		}
		fieldTypes = append(fieldTypes, fieldType)
	}
	return fieldTypes, limit, nil
}

func scanAggregateRow(
	rows execution.Rows, fieldTypes []api.FieldType,
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
	decoded, err := scanner.Decode()
	if err != nil {
		return nil, ErrContract
	}
	return decoded, nil
}
