package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (executor Executor) executeGraph(
	ctx context.Context,
	scope transaction.TenantScope,
	plan resolved.AuthorizedPlan,
) (Result, error) {
	batch, err := executor.postgresCompiler.CompileRecordBatch(plan)
	if err != nil || batch.ValidatePlan(plan) != nil ||
		validateScope(ctx, scope, batch) != nil {
		return Result{}, ErrContract
	}
	consumed, err := consumeBatch(batch, plan)
	if err != nil {
		return Result{}, ErrContract
	}
	graph, err := executor.postgresCompiler.CompileIncludeGraph(plan)
	if err != nil || graph.ValidatePlan(plan) != nil {
		return Result{}, ErrContract
	}
	rootIDAt, err := rootStableColumn(plan, consumed.record)
	if err != nil {
		return Result{}, ErrContract
	}
	runtime, err := newGraphRuntime(
		plan, graph, executor.postgresCompiler, consumed,
		executor.codec, executor.policy, rootIDAt,
	)
	if err != nil {
		return Result{}, ErrContract
	}
	programStage, err := runtime.programs(ctx, scope)
	if err != nil {
		return Result{}, ErrContract
	}
	programs := startGraphPrograms(executor.registry, programStage)
	barrierErr, programErr := runGraphLifecycle(ctx, runtime.barrier, programs)
	if barrierErr != nil {
		return Result{}, barrierErr
	}
	if programErr != nil {
		return Result{}, programErr
	}
	return runtime.result()
}
