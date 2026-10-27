package systemresource

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	formpresentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
)

func BuildFieldCatalog(
	published catalog.Published,
	visibility CatalogVisibility,
) (publicresource.FieldCatalog, error) {
	if published.Validate() != nil || visibility.Validate() != nil {
		return publicresource.FieldCatalog{}, publicresource.ErrContract
	}
	resource, err := publicResource(published)
	if err != nil {
		return publicresource.FieldCatalog{}, err
	}
	set, err := published.CatalogSet()
	if err != nil {
		return publicresource.FieldCatalog{}, publicresource.ErrContract
	}
	root, err := set.RootCatalog()
	if err != nil {
		return publicresource.FieldCatalog{}, publicresource.ErrContract
	}
	definitions, err := root.Fields()
	if err != nil {
		return publicresource.FieldCatalog{}, publicresource.ErrContract
	}
	fields := make([]formpresentation.Field, 0, len(definitions))
	descriptors := make([]publicresource.FieldDescriptor, 0, len(definitions))
	for _, definition := range definitions {
		field, include, fieldErr := presentationField(
			published, resource, definition, visibility,
		)
		if fieldErr != nil {
			return publicresource.FieldCatalog{}, fieldErr
		}
		if include {
			fields = append(fields, field)
			descriptor, descriptorErr := publicFieldDescriptor(published, definition)
			if descriptorErr != nil {
				return publicresource.FieldCatalog{}, descriptorErr
			}
			descriptors = append(descriptors, descriptor)
		}
	}
	fields, descriptors, err = appendRecordProjectionCatalog(
		resource, fields, descriptors,
	)
	if err != nil {
		return publicresource.FieldCatalog{}, err
	}
	presentation, err := formpresentation.NewFieldCatalog(fields)
	if err != nil {
		return publicresource.FieldCatalog{}, publicresource.ErrContract
	}
	resourceID, versionID, querySHA, err := catalogReference(published)
	if err != nil {
		return publicresource.FieldCatalog{}, err
	}
	return publicresource.NewFieldCatalog(
		resource, resourceID, versionID, querySHA, presentation, descriptors,
	)
}
