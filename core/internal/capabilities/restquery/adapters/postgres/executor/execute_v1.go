package executor

import (
	"context"

	v1query "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

// ExecuteV1 executes page=1 normally and page>1 through a sealed predecessor seek.
func (executor Executor) ExecuteV1(
	ctx context.Context,
	scope transaction.TenantScope,
	accepted v1query.Accepted,
) (Result, error) {
	if executor.Validate() != nil {
		return Result{}, ErrContract
	}
	plan, pagePlan, acceptErr := accepted.ConsumeExecution()
	ordinal, hasPredecessor, err := validateV1PagePlan(plan, pagePlan)
	if acceptErr != nil || err != nil {
		return Result{}, ErrContract
	}
	if !hasPredecessor {
		return executor.Execute(ctx, scope, plan)
	}
	batch, err := executor.postgresCompiler.CompileRecordBatch(plan)
	if err != nil || batch.ValidatePlan(plan) != nil ||
		validateScope(ctx, scope, batch) != nil {
		return Result{}, ErrContract
	}
	predecessor, err := executor.postgresCompiler.CompileV1Predecessor(plan, ordinal)
	if err != nil || predecessor.ValidatePlan(plan, ordinal) != nil {
		return Result{}, ErrContract
	}
	consumed, err := consumeBatch(batch, plan)
	if err != nil || !consumed.hasExactCount {
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
			next, runErr := runV1(
				programContext, driver, plan, ordinal, predecessor, consumed,
				executor.postgresCompiler, executor.codec, executor.policy,
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
