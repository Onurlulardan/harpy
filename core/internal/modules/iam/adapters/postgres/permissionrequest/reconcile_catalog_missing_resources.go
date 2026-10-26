package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
)

func (store *permissionRequestStore) ReconcileMissingCatalogResources(
	ctx context.Context,
	scope tenantScope,
	query permissionrequest.MissingCatalogQuery,
) (permissionrequest.CatalogResourcePage, error) {
	if query.Validate() != nil {
		return permissionrequest.CatalogResourcePage{}, permissionrequest.ErrInvalid
	}
	return store.catalogResourcePage(
		ctx, scope, query.CatalogQuery,
		permissionRequestCatalogMissingResourcesSQL, query.ActionKey,
	)
}
