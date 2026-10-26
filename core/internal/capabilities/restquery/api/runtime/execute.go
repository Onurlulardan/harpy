package runtime

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (engine *Engine) Execute(
	ctx context.Context,
	scope transaction.TenantScope,
	resource string,
	request api.QueryRequest,
) (Result, error) {
	if engine == nil || nilInterface(engine.catalogs) || ctx == nil || resource == "" ||
		request.Validate() != nil {
		return Result{}, apiRuntimeError()
	}
	catalogs, err := engine.catalogs.Catalog(ctx, scope, resource)
	if err != nil {
		return Result{}, err
	}
	return engine.ExecuteCatalog(ctx, scope, catalogs, request)
}

func apiRuntimeError() error {
	return &runtimeContractError{}
}

type runtimeContractError struct{}

func (*runtimeContractError) Error() string { return "REST Query runtime contract failed" }
