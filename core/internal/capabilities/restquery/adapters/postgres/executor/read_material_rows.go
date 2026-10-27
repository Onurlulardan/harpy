package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/rowdecode"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func readMaterialRows(
	ctx context.Context,
	driver execution.SnapshotDriver,
	policy planner.Policy,
	candidate statement,
	fieldTypes []api.FieldType,
	maximum int,
) ([][]rowdecode.Value, error) {
	if ctx == nil || interfaceIsNil(driver) || len(fieldTypes) < 1 ||
		maximum < 1 || candidate.class != planner.StatementClassRecord &&
		candidate.class != planner.StatementClassInclude {
		return nil, ErrContract
	}
	if err := applyTimeout(ctx, driver, policy, candidate.class); err != nil {
		return nil, err
	}
	rows, err := driver.Query(ctx, candidate.sql, candidate.arguments...)
	if err != nil {
		return nil, mapDatabaseError(ctx, policy, candidate.class, err)
	}
	if interfaceIsNil(rows) {
		return nil, ErrContract
	}
	defer rows.Close()
	decoded := make([][]rowdecode.Value, 0, min(maximum, 1024))
	for rows.Next() {
		if len(decoded) == maximum {
			return nil, ErrContract
		}
		row, err := scanRecordRow(rows, fieldTypes)
		if err != nil {
			return nil, mapScanError(ctx, policy, candidate.class, err)
		}
		decoded = append(decoded, row)
	}
	if err := rows.Err(); err != nil {
		return nil, mapDatabaseError(ctx, policy, candidate.class, err)
	}
	return decoded, nil
}
