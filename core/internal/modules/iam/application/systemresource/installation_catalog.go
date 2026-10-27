package systemresource

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	"github.com/google/uuid"
)

func BuildInstallationFieldCatalog(
	schemaID, fieldID uuid.UUID,
	resource publicresource.Resource,
) (publicresource.FieldCatalog, error) {
	profileResource, err := installationProfileResource(resource)
	if err != nil {
		return publicresource.FieldCatalog{}, err
	}
	field, err := profile.NewField(profile.FieldInput{
		ID: fieldID, Resource: profileResource, Key: "display_name",
		Type: profile.LogicalString, Cardinality: profile.CardinalitySingle,
		Required: true, Classification: profile.ClassificationPublic,
		Normalization: profile.NormalizationNoneV1, State: profile.FieldActive,
		Validation: profile.ValidationInput{MaximumBytes: 128},
	})
	if err != nil {
		return publicresource.FieldCatalog{}, publicresource.ErrContract
	}
	schema, err := profile.NewPublishedSchema(profile.SchemaInput{
		ID: schemaID, Resource: profileResource, Version: 1,
		State: profile.SchemaActive, Fields: []profile.Field{field},
	})
	if err != nil {
		return publicresource.FieldCatalog{}, publicresource.ErrContract
	}
	custom, err := profile.NewQueryCatalog(schema)
	if err != nil {
		return publicresource.FieldCatalog{}, publicresource.ErrContract
	}
	published, err := catalog.Build(custom)
	if err != nil {
		return publicresource.FieldCatalog{}, publicresource.ErrContract
	}
	visibility, err := NewCatalogVisibility(true, true, true)
	if err != nil {
		return publicresource.FieldCatalog{}, publicresource.ErrContract
	}
	fieldCatalog, err := BuildFieldCatalog(published, visibility)
	if err != nil {
		return publicresource.FieldCatalog{}, err
	}
	if resource == publicresource.Organizations {
		return installationOrganizationLayoutCatalog(fieldCatalog)
	}
	return fieldCatalog, nil
}

func installationProfileResource(
	resource publicresource.Resource,
) (profile.ResourceKind, error) {
	switch resource {
	case publicresource.Users:
		return profile.ResourceUser, nil
	case publicresource.Organizations:
		return profile.ResourceOrganization, nil
	case publicresource.Roles:
		return profile.ResourceRole, nil
	default:
		return "", publicresource.ErrContract
	}
}
