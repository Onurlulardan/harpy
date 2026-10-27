package systemresource

import (
	restquery "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
)

func publicFieldDescriptor(
	published catalog.Published,
	definition restquery.FieldDefinition,
) (publicresource.FieldDescriptor, error) {
	id, idErr := definition.ID()
	name, nameErr := definition.PublicName()
	capabilities, capabilitiesErr := definition.Capabilities()
	if idErr != nil || nameErr != nil || capabilitiesErr != nil {
		return publicresource.FieldDescriptor{}, publicresource.ErrContract
	}
	sortable, sortableErr := capabilities.Sortable()
	searchable, searchableErr := capabilities.Searchable()
	required, requiredErr := published.RequiredForCreate(id)
	maximumBytes, maximumErr := published.MaximumInputBytes(id)
	if sortableErr != nil || searchableErr != nil || requiredErr != nil ||
		maximumErr != nil {
		return publicresource.FieldDescriptor{}, publicresource.ErrContract
	}
	descriptor, err := publicresource.NewFieldDescriptorWithContract(
		publicresource.ID(id), name, sortable, searchable, required, maximumBytes,
	)
	if err != nil {
		return publicresource.FieldDescriptor{}, publicresource.ErrContract
	}
	return descriptor, nil
}
