package executor

import (
	"context"
	"strconv"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func applyTimeout(
	ctx context.Context,
	driver execution.QueryDriver,
	policy planner.Policy,
	class planner.StatementClass,
) error {
	statementPolicy, err := policy.Statement(class)
	if err != nil {
		return ErrContract
	}
	value, err := timeoutValue(statementPolicy.Timeout())
	if err != nil {
		return err
	}
	if err := driver.Execute(
		ctx, compiler.TransactionTimeoutSQL(), value,
	); err != nil {
		return mapDatabaseError(ctx, policy, class, err)
	}
	return nil
}

func timeoutValue(timeout time.Duration) (string, error) {
	if timeout <= 0 || timeout%time.Millisecond != 0 {
		return "", ErrContract
	}
	return strconv.FormatInt(timeout.Milliseconds(), 10) + "ms", nil
}
