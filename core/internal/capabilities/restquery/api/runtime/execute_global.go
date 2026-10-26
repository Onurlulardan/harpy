package runtime

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolver"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (engine *GlobalEngine) Execute(
	ctx context.Context,
	scope transaction.GlobalScope,
	resource string,
	request api.QueryRequest,
) (Result, error) {
	if engine == nil || ctx == nil || resource == "" || request.Validate() != nil ||
		request.Mode() != api.QueryModeRecord || len(request.Includes()) != 0 {
		return Result{}, apiRuntimeError()
	}
	snapshotAt, err := engine.snapshotAt(request)
	if err != nil {
		return Result{}, err
	}
	catalogs, err := engine.catalogs.Catalog(ctx, scope, resource)
	if err != nil {
		return Result{}, err
	}
	spec, err := resolver.Resolve(request, catalogs)
	if err != nil {
		return Result{}, err
	}
	plan, err := engine.authorizer.Authorize(ctx, scope, spec, snapshotAt)
	if err != nil {
		return Result{}, err
	}
	result, err := engine.executor.Execute(ctx, scope, plan)
	if err != nil {
		return Result{}, err
	}
	return recordResult(result)
}

func (engine *GlobalEngine) snapshotAt(request api.QueryRequest) (time.Time, error) {
	after, found := request.Page().After()
	if found {
		return engine.codec.AuthenticatedGlobalSnapshotAt(after)
	}
	return engine.codec.SnapshotNow()
}
