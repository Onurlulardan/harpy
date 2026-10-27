package application

import (
	"context"
	"errors"

	presentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func (application *PresentationApplication) Resolve(
	ctx context.Context,
	resource layoutapi.ResourceKind,
	catalog presentation.FieldCatalog,
) (presentation.Document, error) {
	if !application.usable(ctx) || !resource.Valid() || catalog.Validate() != nil {
		return presentation.Document{}, layoutapi.ErrInvalid
	}
	published, err := application.store.Resolve(ctx, resource)
	if errors.Is(err, layoutapi.ErrNotFound) {
		return defaultPresentation(resource, catalog)
	}
	if err != nil {
		return presentation.Document{}, err
	}
	return verifyPublishedPresentation(published, resource, catalog)
}
