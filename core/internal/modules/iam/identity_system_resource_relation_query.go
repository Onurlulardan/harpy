package iam

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	"github.com/google/uuid"
)

const identityRelationQueryBatchSize = queryapp.RelationLookupBatchSize

func newIdentityRelationQuery(
	ids []uuid.UUID,
	fields []string,
) (publicresource.QueryRequest, error) {
	request, err := queryapp.NewRelationLookupRequest(ids, fields)
	if err != nil {
		return nil, publicresource.ErrContract
	}
	return request, nil
}

func identityRelationSourceFields(
	resource publicresource.Resource,
	catalog publicresource.FieldCatalog,
) ([]string, error) {
	descriptors, err := catalog.Descriptors()
	if err != nil {
		return nil, publicresource.ErrUnavailable
	}
	available := make(map[string]struct{}, len(descriptors))
	for _, descriptor := range descriptors {
		name, nameErr := descriptor.PublicName()
		if nameErr != nil {
			return nil, publicresource.ErrUnavailable
		}
		available[name] = struct{}{}
	}
	result := []string{"id"}
	for _, definition := range identityRelationDefinitions(resource) {
		if _, found := available[definition.field]; found {
			result = append(result, definition.field)
		}
	}
	return result, nil
}
