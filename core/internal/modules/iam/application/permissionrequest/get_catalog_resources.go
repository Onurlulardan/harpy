package permissionrequest

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *CatalogService) GetResources(
	ctx context.Context,
	scope transaction.TenantScope,
	query CatalogQuery,
) (CatalogResourcePage, error) {
	if service == nil || service.store == nil || query.Validate() != nil ||
		scope.ActorID() == uuid.Nil {
		return CatalogResourcePage{}, ErrInvalid
	}
	page, err := service.store.GetCatalogResources(ctx, scope, query)
	if err != nil {
		return CatalogResourcePage{}, err
	}
	return acceptedCatalogResourcePage(page, query.Limit)
}

func acceptedCatalogResourcePage(
	page CatalogResourcePage,
	limit int,
) (CatalogResourcePage, error) {
	if len(page.Items) > limit {
		return CatalogResourcePage{}, ErrUnavailable
	}
	for _, resource := range page.Items {
		if resource.Validate() != nil {
			return CatalogResourcePage{}, ErrUnavailable
		}
	}
	return page, nil
}
