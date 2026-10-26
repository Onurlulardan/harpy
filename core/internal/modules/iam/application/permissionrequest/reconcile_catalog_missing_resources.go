package permissionrequest

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *CatalogService) ReconcileMissingResources(
	ctx context.Context,
	scope transaction.TenantScope,
	query MissingCatalogQuery,
) (CatalogResourcePage, error) {
	if service == nil || service.store == nil || query.Validate() != nil ||
		scope.ActorID() == uuid.Nil {
		return CatalogResourcePage{}, ErrInvalid
	}
	page, err := service.store.ReconcileMissingCatalogResources(ctx, scope, query)
	if err != nil {
		return CatalogResourcePage{}, err
	}
	return acceptedCatalogResourcePage(page, query.Limit)
}
