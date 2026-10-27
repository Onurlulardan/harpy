package systemresource

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	formpresentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
)

func installationOrganizationLayoutCatalog(
	catalog publicresource.FieldCatalog,
) (publicresource.FieldCatalog, error) {
	resource, err := catalog.Resource()
	if err != nil || resource != publicresource.Organizations {
		return publicresource.FieldCatalog{}, publicresource.ErrContract
	}
	resourceID, versionID, querySHA, err := catalog.Reference()
	if err != nil {
		return publicresource.FieldCatalog{}, publicresource.ErrContract
	}
	presentationCatalog, err := catalog.Presentation()
	if err != nil {
		return publicresource.FieldCatalog{}, publicresource.ErrContract
	}
	fields, err := presentationCatalog.Fields()
	if err != nil {
		return publicresource.FieldCatalog{}, publicresource.ErrContract
	}
	filteredFields := make([]formpresentation.Field, 0, len(fields))
	ownerRemoved := false
	for _, field := range fields {
		id, fieldErr := field.ID()
		if fieldErr != nil {
			return publicresource.FieldCatalog{}, publicresource.ErrContract
		}
		if id == publicresource.OrganizationOwnerIDsFieldID {
			ownerRemoved = true
			continue
		}
		filteredFields = append(filteredFields, field)
	}
	descriptors, err := catalog.Descriptors()
	if err != nil {
		return publicresource.FieldCatalog{}, publicresource.ErrContract
	}
	filteredDescriptors := make([]publicresource.FieldDescriptor, 0, len(descriptors))
	for _, descriptor := range descriptors {
		id, descriptorErr := descriptor.ID()
		if descriptorErr != nil {
			return publicresource.FieldCatalog{}, publicresource.ErrContract
		}
		if id.String() != publicresource.OrganizationOwnerIDsFieldID {
			filteredDescriptors = append(filteredDescriptors, descriptor)
		}
	}
	if !ownerRemoved || len(filteredFields) != len(filteredDescriptors) {
		return publicresource.FieldCatalog{}, publicresource.ErrContract
	}
	compatible, err := formpresentation.NewFieldCatalog(filteredFields)
	if err != nil {
		return publicresource.FieldCatalog{}, publicresource.ErrContract
	}
	return publicresource.NewFieldCatalog(
		resource, resourceID, versionID, querySHA,
		compatible, filteredDescriptors,
	)
}
