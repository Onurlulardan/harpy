package application

import (
	presentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func InstallationDefaultDefinition(
	resource layoutapi.ResourceKind,
	catalog presentation.FieldCatalog,
) (layoutapi.Definition, error) {
	if !resource.Valid() || catalog.Validate() != nil {
		return layoutapi.Definition{}, layoutapi.ErrInvalid
	}
	return defaultPresentationDefinition(resource, catalog)
}
