package permissionrequest

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type CatalogStore interface {
	GetCatalogResources(
		context.Context,
		transaction.TenantScope,
		CatalogQuery,
	) (CatalogResourcePage, error)
	ResolveCatalogActions(
		context.Context,
		transaction.TenantScope,
		CatalogQuery,
	) (CatalogActionPage, error)
	ReconcileMissingCatalogResources(
		context.Context,
		transaction.TenantScope,
		MissingCatalogQuery,
	) (CatalogResourcePage, error)
}

type CatalogService struct{ store CatalogStore }

func NewCatalogService(store CatalogStore) (*CatalogService, error) {
	if store == nil {
		return nil, ErrUnavailable
	}
	return &CatalogService{store: store}, nil
}
