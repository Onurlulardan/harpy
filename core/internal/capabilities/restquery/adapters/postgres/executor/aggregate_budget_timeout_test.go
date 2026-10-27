package executor

import (
	"context"
	"errors"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func TestAggregateBudgetRejectsBeforeDataStatement(t *testing.T) {
	fixture := newAggregateFixture(t, true)
	batch := consumedAggregateFixture(t, fixture)
	driver, candidate := readyAggregateDriver(t, batch, nil, "2")
	_, err := runAggregate(context.Background(), driver, batch, aggregatePolicy(t, 1))
	assertAggregateQueryCode(t, err, restqueryv1.ErrorQueryBudgetExceeded)
	if eventIndex(driver.events, "query:"+candidate.sql) >= 0 {
		t.Fatalf("data statement ran after budget rejection: %v", driver.events)
	}
}

func TestAggregateTimeoutOverflowAndCancellationMapping(t *testing.T) {
	fixture := newAggregateFixture(t, true)
	batch := consumedAggregateFixture(t, fixture)
	for name, state := range map[string]struct {
		sqlState string
		code     restqueryv1.ErrorCode
	}{
		"timeout":  {"57014", restqueryv1.ErrorQueryTimeout},
		"overflow": {"22003", restqueryv1.ErrorQueryAggregateOverflow},
	} {
		t.Run(name, func(t *testing.T) {
			driver, candidate := readyAggregateDriver(t, batch, nil, "1")
			driver.queryErrors[candidate.sql] = fakeSQLStateError{state: state.sqlState}
			_, err := runAggregate(
				context.Background(), driver, batch, planner.HardPolicy(),
			)
			assertAggregateQueryCode(t, err, state.code)
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	driver, candidate := readyAggregateDriver(t, batch, nil, "1")
	driver.queryErrors[candidate.sql] = fakeSQLStateError{state: "57014"}
	driver.onQuery = func(statement string) {
		if statement == candidate.sql {
			cancel()
		}
	}
	_, err := runAggregate(ctx, driver, batch, planner.HardPolicy())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error=%v", err)
	}
}

func assertAggregateQueryCode(
	t *testing.T, err error, expected restqueryv1.ErrorCode,
) {
	t.Helper()
	var queryError api.QueryError
	if !errors.As(err, &queryError) || queryError.Code() != expected {
		t.Fatalf("query error=%v expected=%s", err, expected)
	}
}
