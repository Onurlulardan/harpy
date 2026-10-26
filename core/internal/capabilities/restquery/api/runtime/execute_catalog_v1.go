package runtime

import (
	"context"

	v1query "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolver"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type v1CatalogAuthorizer struct {
	engine   *Engine
	scope    transaction.TenantScope
	catalogs api.CatalogSet
}

func (authorizer v1CatalogAuthorizer) ResolveAndAuthorizeV1(
	ctx context.Context, request api.QueryRequest,
) (resolved.AuthorizedPlan, error) {
	if authorizer.engine == nil || request.Validate() != nil ||
		authorizer.catalogs.Validate() != nil {
		return resolved.AuthorizedPlan{}, apiRuntimeError()
	}
	spec, err := resolver.Resolve(request, authorizer.catalogs)
	if err != nil {
		return resolved.AuthorizedPlan{}, err
	}
	return authorizer.engine.authorizer.Authorize(ctx, authorizer.scope, spec)
}

// ExecuteCatalogV1 preserves V1 page semantics while reusing the canonical
// resolver, authorizer, compiler and owner executor inside one snapshot.
func (engine *Engine) ExecuteCatalogV1(
	ctx context.Context,
	scope transaction.TenantScope,
	catalogs api.CatalogSet,
	prepared v1query.Prepared,
) (Result, string, bool, error) {
	if engine == nil || ctx == nil || catalogs.Validate() != nil ||
		prepared.Validate() != nil {
		return Result{}, "", false, apiRuntimeError()
	}
	accepted, err := prepared.Accept(ctx, v1CatalogAuthorizer{
		engine: engine, scope: scope, catalogs: catalogs,
	})
	if err != nil {
		return Result{}, "", false, err
	}
	warning, deprecated, err := accepted.Warning()
	if err != nil {
		return Result{}, "", false, apiRuntimeError()
	}
	executed, err := engine.executor.ExecuteV1(ctx, scope, accepted)
	if err != nil {
		return Result{}, "", false, err
	}
	result, err := recordResult(executed)
	if err != nil {
		return Result{}, "", false, err
	}
	return result, warning, deprecated, nil
}
