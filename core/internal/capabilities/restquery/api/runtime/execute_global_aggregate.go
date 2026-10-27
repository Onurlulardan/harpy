package runtime

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolver"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (engine *GlobalEngine) ExecuteAggregate(
	ctx context.Context,
	scope transaction.GlobalScope,
	resource string,
	request api.QueryRequest,
) (AggregateResult, error) {
	if engine == nil || ctx == nil || resource == "" || request.Validate() != nil ||
		request.Mode() != api.QueryModeAggregate || len(request.Includes()) != 0 {
		return AggregateResult{}, apiRuntimeError()
	}
	snapshotAt, err := engine.codec.SnapshotNow()
	if err != nil {
		return AggregateResult{}, err
	}
	catalogs, err := engine.catalogs.Catalog(ctx, scope, resource)
	if err != nil {
		return AggregateResult{}, err
	}
	spec, err := resolver.Resolve(request, catalogs)
	if err != nil {
		return AggregateResult{}, err
	}
	plan, err := engine.authorizer.Authorize(ctx, scope, spec, snapshotAt)
	if err != nil {
		return AggregateResult{}, err
	}
	result, err := engine.executor.ExecuteAggregate(ctx, scope, plan)
	if err != nil {
		return AggregateResult{}, err
	}
	groups, err := result.Groups()
	limit, limitErr := result.Limit()
	if err != nil || limitErr != nil {
		return AggregateResult{}, apiRuntimeError()
	}
	return newAggregateResult(groups, limit), nil
}
