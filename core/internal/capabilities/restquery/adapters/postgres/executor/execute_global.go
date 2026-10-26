package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (executor GlobalExecutor) Execute(
	ctx context.Context,
	scope transaction.GlobalScope,
	plan resolved.AuthorizedPlan,
) (Result, error) {
	if executor.Validate() != nil || plan.Validate() != nil {
		return Result{}, ErrContract
	}
	spec, err := plan.Spec()
	if err != nil {
		return Result{}, ErrContract
	}
	mode, err := spec.Mode()
	if err != nil || mode != api.QueryModeRecord {
		return Result{}, ErrContract
	}
	hasIncludes, err := planHasIncludes(plan)
	if err != nil || hasIncludes {
		return Result{}, ErrContract
	}
	batch, err := executor.postgresCompiler.CompileRecordBatch(plan)
	if err != nil || batch.ValidatePlan(plan) != nil ||
		validateGlobalScope(ctx, scope, batch, plan) != nil {
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
	program, err := mintGlobalProgram(
		ctx, scope, catalog,
		func(programContext context.Context, driver execution.GlobalSnapshotDriver) error {
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
