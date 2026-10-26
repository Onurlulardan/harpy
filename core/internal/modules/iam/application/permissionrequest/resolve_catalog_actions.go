package permissionrequest

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *CatalogService) ResolveActions(
	ctx context.Context,
	scope transaction.TenantScope,
	query CatalogQuery,
) (CatalogActionPage, error) {
	if service == nil || service.store == nil || query.Validate() != nil ||
		query.Kind != CatalogKindAll ||
		scope.ActorID() == uuid.Nil {
		return CatalogActionPage{}, ErrInvalid
	}
	page, err := service.store.ResolveCatalogActions(ctx, scope, query)
	if err != nil {
		return CatalogActionPage{}, err
	}
	if len(page.Items) > query.Limit {
		return CatalogActionPage{}, ErrUnavailable
	}
	for _, action := range page.Items {
		if action.Validate() != nil {
			return CatalogActionPage{}, ErrUnavailable
		}
	}
	return page, nil
}
