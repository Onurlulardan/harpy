package runtime

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolver"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

// ExecuteCatalog runs the canonical REST Query pipeline against an exact,
// immutable owner-resolved catalog set.
func (engine *Engine) ExecuteCatalog(
	ctx context.Context,
	scope transaction.TenantScope,
	catalogs api.CatalogSet,
	request api.QueryRequest,
) (Result, error) {
	if engine == nil || ctx == nil || catalogs.Validate() != nil || request.Validate() != nil {
		return Result{}, apiRuntimeError()
	}
	spec, err := resolver.Resolve(request, catalogs)
	if err != nil {
		return Result{}, err
	}
	plan, err := engine.authorizer.Authorize(ctx, scope, spec)
	if err != nil {
		return Result{}, err
	}
	if request.Mode() == api.QueryModeAggregate {
		aggregate, executeErr := engine.executor.ExecuteAggregate(ctx, scope, plan)
		if executeErr != nil {
			return Result{}, executeErr
		}
		items, itemsErr := aggregate.Groups()
		if itemsErr != nil {
			return Result{}, apiRuntimeError()
		}
		return newResult(items, false, "", nil), nil
	}
	result, err := engine.executor.Execute(ctx, scope, plan)
	if err != nil {
		return Result{}, err
	}
	return recordResult(result)
}
