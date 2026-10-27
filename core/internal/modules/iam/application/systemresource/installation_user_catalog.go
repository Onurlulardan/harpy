package systemresource

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	"github.com/google/uuid"
)

// BuildInstallationUserFieldCatalog keeps the bootstrap display name while
// publishing the aliases required by the identity HTTP contract.
func BuildInstallationUserFieldCatalog(
	schemaID, displayID, firstNameID, lastNameID, phoneID uuid.UUID,
) (publicresource.FieldCatalog, error) {
	fields := make([]profile.Field, 0, 4)
	for _, input := range []profile.FieldInput{
		installationUserField(displayID, "display_name", true, 128),
		installationUserField(firstNameID, "first_name", false, 200),
		installationUserField(lastNameID, "last_name", false, 200),
		installationUserField(phoneID, "phone", false, 32),
	} {
		field, err := profile.NewField(input)
		if err != nil {
			return publicresource.FieldCatalog{}, publicresource.ErrContract
		}
		fields = append(fields, field)
	}
	schema, err := profile.NewPublishedSchema(profile.SchemaInput{
		ID: schemaID, Resource: profile.ResourceUser, Version: 2,
		State: profile.SchemaActive, Fields: fields,
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
	return BuildFieldCatalog(published, visibility)
}

func installationUserField(
	id uuid.UUID, key string, required bool, maximumBytes int,
) profile.FieldInput {
	return profile.FieldInput{
		ID: id, Resource: profile.ResourceUser, Key: key,
		Type: profile.LogicalString, Cardinality: profile.CardinalitySingle,
		Required: required, Classification: profile.ClassificationPublic,
		Normalization: profile.NormalizationNoneV1, State: profile.FieldActive,
		Validation: profile.ValidationInput{MaximumBytes: maximumBytes},
	}
}
