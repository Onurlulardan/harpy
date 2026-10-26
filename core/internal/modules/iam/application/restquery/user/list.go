package user

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *Service) List(
	ctx context.Context,
	scope transaction.TenantScope,
) (catalog.Published, error) {
	if service == nil || portIsNil(service.port) || !scopeUsable(ctx, scope) {
		return catalog.Published{}, catalog.ErrContract
	}
	published, err := service.port.List(ctx, scope)
	if err != nil || published.Validate() != nil {
		return catalog.Published{}, catalog.ErrContract
	}
	resource, err := published.Resource()
	if err != nil || resource.Kind() != profile.ResourceUser {
		return catalog.Published{}, catalog.ErrContract
	}
	return published, nil
}
