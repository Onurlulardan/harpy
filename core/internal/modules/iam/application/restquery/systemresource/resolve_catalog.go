package systemresourcequery

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *Service) ResolveCatalog(
	ctx context.Context,
	scope transaction.TenantScope,
	resource string,
) (catalog.Published, error) {
	if service == nil || service.engine == nil || ctx == nil {
		return catalog.Published{}, ErrContract
	}
	var (
		published catalog.Published
		err       error
	)
	switch resource {
	case "users":
		published, err = service.catalogs.users.List(ctx, scope)
	case "roles":
		published, err = service.catalogs.roles.List(ctx, scope)
	case "organizations":
		published, err = service.catalogs.organizations.List(ctx, scope)
	default:
		return catalog.Published{}, ErrContract
	}
	if err != nil {
		return catalog.Published{}, NormalizeError(err)
	}
	if published.Validate() != nil {
		return catalog.Published{}, ErrContract
	}
	return published, nil
}
