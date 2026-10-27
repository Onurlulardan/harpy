package systemresources

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Catalog(
	ctx context.Context,
	admission Admission,
	resource publicresource.Resource,
) (CatalogResult, error) {
	if !resource.Valid() {
		return CatalogResult{}, ErrContract
	}
	actor, err := admission.Actor()
	if err != nil {
		return CatalogResult{}, err
	}
	var result CatalogResult
	err = application.runSnapshot(
		ctx, admission, func(scopedContext context.Context, scope platformtx.TenantScope) error {
			catalog, scopedErr := application.iam.Catalog(
				scopedContext, scope, actor, resource,
			)
			if scopedErr != nil {
				return scopedErr
			}
			presentation, scopedErr := application.resolvePresentation(
				scopedContext, scope, catalog,
			)
			if scopedErr != nil {
				return scopedErr
			}
			result, scopedErr = NewCatalogResult(presentation)
			return scopedErr
		},
	)
	if err != nil {
		return CatalogResult{}, err
	}
	return result, nil
}
