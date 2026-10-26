package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (executor GlobalExecutor) ExecuteAggregate(
	ctx context.Context,
	scope transaction.GlobalScope,
	plan resolved.AuthorizedPlan,
) (AggregateResult, error) {
	if executor.Validate() != nil || plan.Validate() != nil {
		return AggregateResult{}, ErrContract
	}
	batch, err := executor.postgresCompiler.CompileAggregateBatch(plan)
	if err != nil || batch.ValidatePlan(plan) != nil ||
		validateGlobalAggregateScope(ctx, scope, batch, plan) != nil {
		return AggregateResult{}, ErrContract
	}
	consumed, err := consumeAggregateBatch(batch, plan)
	if err != nil {
		return AggregateResult{}, ErrContract
	}
	catalog, err := batch.Catalog()
	if err != nil {
		return AggregateResult{}, ErrContract
	}
	var result AggregateResult
	program, err := mintGlobalProgram(
		ctx, scope, catalog,
		func(programContext context.Context, driver execution.GlobalSnapshotDriver) error {
			next, runErr := runAggregate(programContext, driver, consumed, executor.policy)
			if runErr == nil {
				result = next
			}
			return runErr
		},
	)
	if err != nil {
		return AggregateResult{}, ErrContract
	}
	if err := executor.registry.Run(program); err != nil {
		return AggregateResult{}, err
	}
	if result.Validate() != nil {
		return AggregateResult{}, ErrContract
	}
	return result, nil
}
