package executor

import (
	"context"
	"errors"
	"reflect"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

const queryCanceledSQLState = "57014"
const numericOutOfRangeSQLState = "22003"

type sqlStateError interface {
	SQLState() string
}

func mapDatabaseError(
	ctx context.Context,
	policy planner.Policy,
	class planner.StatementClass,
	err error,
) error {
	if ctx != nil && ctx.Err() != nil {
		return ctx.Err()
	}
	var state sqlStateError
	if errors.As(err, &state) && state.SQLState() == numericOutOfRangeSQLState &&
		class == planner.StatementClassAggregate {
		return aggregateOverflowError()
	}
	if errors.As(err, &state) && state.SQLState() == queryCanceledSQLState {
		statementPolicy, policyErr := policy.Statement(class)
		if policyErr != nil {
			return ErrContract
		}
		return statementPolicy.TimeoutError()
	}
	return ErrExecution
}

func aggregateOverflowError() error {
	queryError, err := api.NewQueryError(
		restqueryv1.ErrorQueryAggregateOverflow, nil,
	)
	if err != nil {
		return ErrContract
	}
	return queryError
}

func mapScanError(
	ctx context.Context,
	policy planner.Policy,
	class planner.StatementClass,
	err error,
) error {
	if ctx != nil && ctx.Err() != nil {
		return ctx.Err()
	}
	var state sqlStateError
	if errors.As(err, &state) {
		return mapDatabaseError(ctx, policy, class, err)
	}
	return ErrContract
}

func interfaceIsNil(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface,
		reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
