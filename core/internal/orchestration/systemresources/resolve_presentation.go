package systemresources

import (
	"context"
	"errors"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) resolvePresentation(
	ctx context.Context,
	scope platformtx.TenantScope,
	catalog publicresource.FieldCatalog,
) (Presentation, error) {
	resource, err := catalog.Resource()
	if err != nil {
		return Presentation{}, ErrContract
	}
	kind, err := layoutResource(resource)
	if err != nil {
		return Presentation{}, err
	}
	fields, err := catalog.Presentation()
	if err != nil {
		return Presentation{}, ErrContract
	}
	document, err := application.layout.ResolvePresentation(ctx, scope, kind, fields)
	if err != nil {
		return Presentation{}, err
	}
	return NewPresentation(catalog, document)
}

func (application *Application) recordPresentation(
	ctx context.Context,
	scope platformtx.TenantScope,
	catalog publicresource.FieldCatalog,
) (Presentation, bool, error) {
	presentation, err := application.resolvePresentation(ctx, scope, catalog)
	if err == nil {
		return presentation, false, nil
	}
	if errors.Is(err, layoutapi.ErrUnavailable) ||
		errors.Is(err, layoutapi.ErrNotFound) ||
		errors.Is(err, layoutapi.ErrInvalid) {
		return Presentation{}, true, nil
	}
	return Presentation{}, false, err
}
