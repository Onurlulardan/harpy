package systemresourcequery

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *Service) Execute(
	ctx context.Context,
	scope transaction.TenantScope,
	resource string,
	request Request,
) (Result, error) {
	published, err := service.ResolveCatalog(ctx, scope, resource)
	if err != nil {
		return Result{}, ErrContract
	}
	return service.RunCatalog(ctx, scope, published, request)
}
