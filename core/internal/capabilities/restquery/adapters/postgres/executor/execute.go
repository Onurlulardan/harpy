package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (executor Executor) Execute(
	ctx context.Context,
	scope transaction.TenantScope,
	plan resolved.AuthorizedPlan,
) (Result, error) {
	if executor.Validate() != nil || plan.Validate() != nil {
		return Result{}, ErrContract
	}
	hasIncludes, err := planHasIncludes(plan)
	if err != nil {
		return Result{}, ErrContract
	}
	if hasIncludes {
		return executor.executeGraph(ctx, scope, plan)
	}
	batch, err := executor.postgresCompiler.CompileRecordBatch(plan)
	if err != nil || batch.ValidatePlan(plan) != nil ||
		validateScope(ctx, scope, batch) != nil {
		return Result{}, ErrContract
	}
	consumed, err := consumeBatch(batch, plan)
	if err != nil {
		return Result{}, ErrContract
	}
	catalog, err := batch.Catalog()
	if err != nil {
		return Result{}, ErrContract
	}
	var result Result
	program, err := mintProgram(
		ctx, scope, catalog,
		func(programContext context.Context, driver execution.SnapshotDriver) error {
			next, runErr := runBatch(
				programContext, driver, consumed, executor.codec, executor.policy,
			)
			if runErr == nil {
				result = next
			}
			return runErr
		},
	)
	if err != nil {
		return Result{}, ErrContract
	}
	if err := executor.registry.Run(program); err != nil {
		return Result{}, err
	}
	if result.Validate() != nil {
		return Result{}, ErrContract
	}
	return result, nil
}
