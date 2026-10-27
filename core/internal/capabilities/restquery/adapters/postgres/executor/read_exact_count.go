package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func readExactCount(
	ctx context.Context,
	driver execution.QueryDriver,
	policy planner.Policy,
	statement statement,
) (int64, error) {
	if statement.class != planner.StatementClassExactCount {
		return 0, ErrContract
	}
	if err := applyTimeout(ctx, driver, policy, statement.class); err != nil {
		return 0, err
	}
	row := driver.QueryRow(ctx, statement.sql, statement.arguments...)
	if interfaceIsNil(row) {
		return 0, ErrContract
	}
	var total int64
	if err := row.Scan(&total); err != nil {
		return 0, mapScanError(ctx, policy, statement.class, err)
	}
	if total < 0 {
		return 0, ErrContract
	}
	return total, nil
}
