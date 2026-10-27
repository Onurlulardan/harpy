package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func runAggregate(
	ctx context.Context,
	driver execution.QueryDriver,
	batch consumedAggregate,
	policy planner.Policy,
) (AggregateResult, error) {
	if ctx == nil || ctx.Err() != nil || interfaceIsNil(driver) ||
		batch.Validate() != nil || validateAggregateOwner(driver, batch) != nil {
		return AggregateResult{}, ErrContract
	}
	candidate, err := aggregateStatement(batch.statement)
	if err != nil {
		return AggregateResult{}, err
	}
	if err := admitStatements(ctx, driver, policy, []statement{candidate}); err != nil {
		return AggregateResult{}, err
	}
	page, err := readAggregate(ctx, driver, policy, batch.statement, candidate)
	if err != nil {
		return AggregateResult{}, err
	}
	if page.Validate() != nil {
		return AggregateResult{}, ErrContract
	}
	return newAggregateResult(page)
}

func validateAggregateOwner(
	driver execution.QueryDriver, batch consumedAggregate,
) error {
	owner, err := batch.statement.Owner()
	if err != nil || driver.Owner() != owner {
		return ErrContract
	}
	return nil
}
