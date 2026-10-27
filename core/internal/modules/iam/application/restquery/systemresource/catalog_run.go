package systemresourcequery

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *Service) RunCatalog(
	ctx context.Context,
	scope transaction.TenantScope,
	published catalog.Published,
	request Request,
) (Result, error) {
	value, requestErr := request.Value()
	set, catalogErr := published.CatalogSet()
	if service == nil || service.engine == nil || ctx == nil ||
		requestErr != nil || catalogErr != nil {
		return Result{}, ErrContract
	}
	result, err := service.engine.ExecuteCatalog(ctx, scope, set, value)
	if err != nil {
		return Result{}, NormalizeError(err)
	}
	transport := ""
	if result.HasMore() {
		transport = request.Transport()
	}
	return newResult(result, transport), nil
}
